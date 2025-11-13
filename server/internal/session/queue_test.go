package session

import (
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/rules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQueue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Queue Suite")
}

var _ = Describe("Command Queue", Label("scope:unit", "loop:g3-orch", "layer:sim", "double:fake-io", "b:command-ordering", "r:high"), func() {
	Describe("Queue Creation", func() {
		It("creates queue with max size", func() {
			queue := NewCommandQueue(100)
			Expect(queue.maxSize).To(Equal(100))
			Expect(queue.Size()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
		})
	})

	Describe("Basic Operations", func() {
		It("enqueue adds commands", func() {
			queue := NewCommandQueue(10)
			cmd := rules.InputCommand{Thrust: 1.0, Turn: 0.0}

			success := queue.Enqueue(1, cmd)
			Expect(success).To(BeTrue())
			Expect(queue.Size()).To(Equal(1))
			Expect(queue.IsEmpty()).To(BeFalse())
		})

		It("dequeue retrieves commands", func() {
			queue := NewCommandQueue(10)
			cmd := rules.InputCommand{Thrust: 1.0, Turn: 0.0}

			queue.Enqueue(1, cmd)
			dequeued, ok := queue.Dequeue()

			Expect(ok).To(BeTrue())
			Expect(dequeued).NotTo(BeNil())
			Expect(dequeued.Sequence).To(Equal(uint32(1)))
			Expect(dequeued.Command.Thrust).To(Equal(float32(1.0)))
			Expect(queue.Size()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
		})

		It("dequeue returns false when empty", func() {
			queue := NewCommandQueue(10)

			_, ok := queue.Dequeue()
			Expect(ok).To(BeFalse())
		})

		It("peek returns next command without removing", func() {
			queue := NewCommandQueue(10)
			cmd := rules.InputCommand{Thrust: 0.5, Turn: 0.3}

			queue.Enqueue(1, cmd)
			peeked, ok := queue.Peek()

			Expect(ok).To(BeTrue())
			Expect(peeked).NotTo(BeNil())
			Expect(peeked.Sequence).To(Equal(uint32(1)))
			Expect(peeked.Command.Thrust).To(Equal(float32(0.5)))
			Expect(peeked.Command.Turn).To(Equal(float32(0.3)))
			Expect(queue.Size()).To(Equal(1)) // Size unchanged
		})

		It("peek returns false when empty", func() {
			queue := NewCommandQueue(10)

			_, ok := queue.Peek()
			Expect(ok).To(BeFalse())
		})

		It("clear empties the queue", func() {
			queue := NewCommandQueue(10)
			queue.Enqueue(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			queue.Enqueue(2, rules.InputCommand{Thrust: 0.0, Turn: 1.0})

			Expect(queue.Size()).To(Equal(2))
			queue.Clear()
			Expect(queue.Size()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
		})
	})

	Describe("Sequence-Based Deduplication", func() {
		It("rejects duplicate sequence numbers", func() {
			queue := NewCommandQueue(10)
			cmd1 := rules.InputCommand{Thrust: 1.0, Turn: 0.0}
			cmd2 := rules.InputCommand{Thrust: 0.0, Turn: 1.0}

			success1 := queue.Enqueue(1, cmd1)
			Expect(success1).To(BeTrue())

			success2 := queue.Enqueue(1, cmd2) // Same sequence, different command
			Expect(success2).To(BeFalse())
			Expect(queue.Size()).To(Equal(1))

			// Verify first command is still there
			dequeued, ok := queue.Dequeue()
			Expect(ok).To(BeTrue())
			Expect(dequeued.Command.Thrust).To(Equal(float32(1.0))) // Original command
		})

		It("accepts different sequence numbers", func() {
			queue := NewCommandQueue(10)
			cmd1 := rules.InputCommand{Thrust: 1.0, Turn: 0.0}
			cmd2 := rules.InputCommand{Thrust: 0.0, Turn: 1.0}

			success1 := queue.Enqueue(1, cmd1)
			success2 := queue.Enqueue(2, cmd2)

			Expect(success1).To(BeTrue())
			Expect(success2).To(BeTrue())
			Expect(queue.Size()).To(Equal(2))
		})
	})

	Describe("Command Ordering", func() {
		It("dequeues commands in sequence order (lowest first)", func() {
			queue := NewCommandQueue(10)

			// Enqueue out of order
			queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Enqueue(2, rules.InputCommand{Thrust: 0.2, Turn: 0.0})

			// Dequeue should return in order: 1, 2, 3
			cmd1, ok1 := queue.Dequeue()
			Expect(ok1).To(BeTrue())
			Expect(cmd1.Sequence).To(Equal(uint32(1)))
			Expect(cmd1.Command.Thrust).To(Equal(float32(0.1)))

			cmd2, ok2 := queue.Dequeue()
			Expect(ok2).To(BeTrue())
			Expect(cmd2.Sequence).To(Equal(uint32(2)))
			Expect(cmd2.Command.Thrust).To(Equal(float32(0.2)))

			cmd3, ok3 := queue.Dequeue()
			Expect(ok3).To(BeTrue())
			Expect(cmd3.Sequence).To(Equal(uint32(3)))
			Expect(cmd3.Command.Thrust).To(Equal(float32(0.3)))
		})

		It("maintains order for out-of-order inserts", func() {
			queue := NewCommandQueue(10)

			// Insert in random order
			queue.Enqueue(5, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			queue.Enqueue(2, rules.InputCommand{Thrust: 0.2, Turn: 0.0})
			queue.Enqueue(8, rules.InputCommand{Thrust: 0.8, Turn: 0.0})
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})

			// Verify order
			sequences := []uint32{}
			for !queue.IsEmpty() {
				cmd, _ := queue.Dequeue()
				sequences = append(sequences, cmd.Sequence)
			}

			Expect(sequences).To(Equal([]uint32{1, 2, 3, 5, 8}))
		})

		It("handles sequence gaps correctly", func() {
			queue := NewCommandQueue(10)

			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Enqueue(5, rules.InputCommand{Thrust: 0.5, Turn: 0.0}) // Gap: 2, 3, 4 missing
			queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})

			// Should still return in order: 1, 3, 5
			cmd1, _ := queue.Dequeue()
			Expect(cmd1.Sequence).To(Equal(uint32(1)))

			cmd2, _ := queue.Dequeue()
			Expect(cmd2.Sequence).To(Equal(uint32(3)))

			cmd3, _ := queue.Dequeue()
			Expect(cmd3.Sequence).To(Equal(uint32(5)))
		})
	})

	Describe("Queue Bounds", func() {
		It("enforces max size limit", func() {
			queue := NewCommandQueue(3)

			// Fill queue
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Enqueue(2, rules.InputCommand{Thrust: 0.2, Turn: 0.0})
			queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})

			Expect(queue.Size()).To(Equal(3))

			// Try to add one more - should fail
			success := queue.Enqueue(4, rules.InputCommand{Thrust: 0.4, Turn: 0.0})
			Expect(success).To(BeFalse())
			Expect(queue.Size()).To(Equal(3))
		})

		It("accepts commands after dequeue makes space", func() {
			queue := NewCommandQueue(2)

			// Fill queue
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Enqueue(2, rules.InputCommand{Thrust: 0.2, Turn: 0.0})

			// Try to add - should fail
			success1 := queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})
			Expect(success1).To(BeFalse())

			// Dequeue one
			queue.Dequeue()

			// Now should be able to add
			success2 := queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})
			Expect(success2).To(BeTrue())
			Expect(queue.Size()).To(Equal(2))
		})
	})

	Describe("Sequence Number Handling", func() {
		It("rejects commands with sequence < nextSequence (already processed)", func() {
			queue := NewCommandQueue(10)

			// Enqueue and dequeue sequence 1
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Dequeue()

			// Try to enqueue sequence 1 again - should be rejected
			success := queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			Expect(success).To(BeFalse())
		})

		It("accepts commands with sequence >= nextSequence", func() {
			queue := NewCommandQueue(10)

			// Enqueue sequence 1
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})

			// Should accept sequence 2 (next expected)
			success1 := queue.Enqueue(2, rules.InputCommand{Thrust: 0.2, Turn: 0.0})
			Expect(success1).To(BeTrue())

			// Should accept sequence 5 (future command)
			success2 := queue.Enqueue(5, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			Expect(success2).To(BeTrue())
		})

		It("updates nextSequence correctly after dequeue", func() {
			queue := NewCommandQueue(10)

			// Enqueue sequences 1, 3, 5
			queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})
			queue.Enqueue(5, rules.InputCommand{Thrust: 0.5, Turn: 0.0})

			// Dequeue 1
			queue.Dequeue()

			// Try to enqueue 1 again - should be rejected
			success1 := queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			Expect(success1).To(BeFalse())

			// Dequeue 3
			queue.Dequeue()

			// Try to enqueue 1 or 3 again - should be rejected
			success2 := queue.Enqueue(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			success3 := queue.Enqueue(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})
			Expect(success2).To(BeFalse())
			Expect(success3).To(BeFalse())
		})
	})
})

