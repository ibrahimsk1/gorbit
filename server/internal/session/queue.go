package session

import (
	"sort"

	"github.com/gorbit/orbitalrush/internal/sim/rules"
)

// QueuedCommand represents a command with its sequence number.
type QueuedCommand struct {
	Sequence uint32
	Command  rules.InputCommand
}

// CommandQueue is a queue that stores input commands with sequence numbers,
// maintains ordering, and deduplicates by sequence.
type CommandQueue struct {
	commands     map[uint32]*QueuedCommand // O(1) lookup by sequence
	ordered      []uint32                   // Sorted sequence numbers for ordering
	maxSize      int                        // Maximum queue size
	nextSequence uint32                    // Next expected sequence number
}

// NewCommandQueue creates a new command queue with the specified maximum size.
func NewCommandQueue(maxSize int) *CommandQueue {
	return &CommandQueue{
		commands:     make(map[uint32]*QueuedCommand),
		ordered:      make([]uint32, 0),
		maxSize:      maxSize,
		nextSequence: 1, // Start at sequence 1
	}
}

// Enqueue adds a command to the queue with the specified sequence number.
// Returns false if:
//   - The sequence number already exists (duplicate)
//   - The sequence number is less than nextSequence (already processed)
//   - The queue is full
//
// Returns true on success.
func (q *CommandQueue) Enqueue(seq uint32, cmd rules.InputCommand) bool {
	// Reject if sequence is less than nextSequence (already processed)
	if seq < q.nextSequence {
		return false
	}

	// Reject if sequence already exists (duplicate)
	if _, exists := q.commands[seq]; exists {
		return false
	}

	// Reject if queue is full
	if len(q.commands) >= q.maxSize {
		return false
	}

	// Add command
	queuedCmd := &QueuedCommand{
		Sequence: seq,
		Command:  cmd,
	}
	q.commands[seq] = queuedCmd

	// Insert sequence into ordered slice (maintain sorted order)
	q.ordered = append(q.ordered, seq)
	sort.Slice(q.ordered, func(i, j int) bool {
		return q.ordered[i] < q.ordered[j]
	})

	return true
}

// Dequeue removes and returns the next command in sequence order (lowest sequence first).
// Returns false if the queue is empty.
func (q *CommandQueue) Dequeue() (*QueuedCommand, bool) {
	if len(q.ordered) == 0 {
		return nil, false
	}

	// Get the lowest sequence number
	seq := q.ordered[0]

	// Remove from ordered slice
	q.ordered = q.ordered[1:]

	// Remove from map
	cmd := q.commands[seq]
	delete(q.commands, seq)

	// Update nextSequence to be one more than the dequeued sequence
	q.nextSequence = seq + 1

	return cmd, true
}

// Peek returns the next command without removing it.
// Returns false if the queue is empty.
func (q *CommandQueue) Peek() (*QueuedCommand, bool) {
	if len(q.ordered) == 0 {
		return nil, false
	}

	seq := q.ordered[0]
	cmd := q.commands[seq]
	return cmd, true
}

// Size returns the current number of commands in the queue.
func (q *CommandQueue) Size() int {
	return len(q.commands)
}

// IsEmpty returns true if the queue is empty.
func (q *CommandQueue) IsEmpty() bool {
	return len(q.commands) == 0
}

// Clear removes all commands from the queue.
func (q *CommandQueue) Clear() {
	q.commands = make(map[uint32]*QueuedCommand)
	q.ordered = make([]uint32, 0)
	// Note: nextSequence is not reset, as it tracks what has been processed
}

