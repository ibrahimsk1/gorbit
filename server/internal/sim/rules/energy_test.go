package rules

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEnergy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Energy Economy Suite")
}

var _ = Describe("Energy Economy", Label("scope:unit", "loop:g2-rules", "layer:sim", "dep:none", "b:energy-economy", "r:high", "double:fake"), func() {
	Describe("DrainEnergyOnThrust", func() {
		It("drains energy by ThrustDrainRate when thrusting", func() {
			initialEnergy := float32(100.0)
			newEnergy := DrainEnergyOnThrust(initialEnergy, true)
			Expect(newEnergy).To(BeNumerically("~", initialEnergy-ThrustDrainRate, 0.001))
		})

		It("does not drain energy when not thrusting", func() {
			initialEnergy := float32(75.5)
			newEnergy := DrainEnergyOnThrust(initialEnergy, false)
			Expect(newEnergy).To(Equal(initialEnergy))
		})

		It("cannot drain below zero", func() {
			initialEnergy := float32(0.3)
			newEnergy := DrainEnergyOnThrust(initialEnergy, true)
			Expect(newEnergy).To(Equal(float32(0.0)))
		})

		It("handles draining from exactly zero", func() {
			initialEnergy := float32(0.0)
			newEnergy := DrainEnergyOnThrust(initialEnergy, true)
			Expect(newEnergy).To(Equal(float32(0.0)))
		})

		It("handles multiple sequential drains", func() {
			initialEnergy := float32(100.0)
			energy1 := DrainEnergyOnThrust(initialEnergy, true)
			energy2 := DrainEnergyOnThrust(energy1, true)
			energy3 := DrainEnergyOnThrust(energy2, true)
			expectedEnergy := initialEnergy - 3*ThrustDrainRate
			Expect(energy3).To(BeNumerically("~", expectedEnergy, 0.001))
		})

		It("drains correctly when energy is exactly ThrustDrainRate", func() {
			initialEnergy := ThrustDrainRate
			newEnergy := DrainEnergyOnThrust(initialEnergy, true)
			Expect(newEnergy).To(Equal(float32(0.0)))
		})
	})

	Describe("RestoreEnergyOnPickup", func() {
		It("restores energy by PalletRestoreAmount", func() {
			initialEnergy := float32(50.0)
			newEnergy := RestoreEnergyOnPickup(initialEnergy)
			Expect(newEnergy).To(BeNumerically("~", initialEnergy+PalletRestoreAmount, 0.001))
		})

		It("cannot restore above MaxEnergy", func() {
			initialEnergy := float32(90.0)
			newEnergy := RestoreEnergyOnPickup(initialEnergy)
			Expect(newEnergy).To(Equal(MaxEnergy))
		})

		It("handles restoring from exactly MaxEnergy", func() {
			initialEnergy := MaxEnergy
			newEnergy := RestoreEnergyOnPickup(initialEnergy)
			Expect(newEnergy).To(Equal(MaxEnergy))
		})

		It("handles multiple sequential restores", func() {
			initialEnergy := float32(0.0)
			energy1 := RestoreEnergyOnPickup(initialEnergy)
			energy2 := RestoreEnergyOnPickup(energy1)
			energy3 := RestoreEnergyOnPickup(energy2)
			// After 3 restores, should be at max (3 * 25 = 75, which is < 100)
			expectedEnergy := initialEnergy + 3*PalletRestoreAmount
			Expect(energy3).To(BeNumerically("~", expectedEnergy, 0.001))
		})

		It("restores correctly when energy is near MaxEnergy", func() {
			initialEnergy := MaxEnergy - PalletRestoreAmount + 1.0
			newEnergy := RestoreEnergyOnPickup(initialEnergy)
			Expect(newEnergy).To(Equal(MaxEnergy))
		})

		It("restores from zero energy", func() {
			initialEnergy := float32(0.0)
			newEnergy := RestoreEnergyOnPickup(initialEnergy)
			Expect(newEnergy).To(Equal(PalletRestoreAmount))
		})
	})

	Describe("ClampEnergy", func() {
		It("clamps negative energy to zero", func() {
			energy := float32(-10.0)
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(float32(0.0)))
		})

		It("clamps energy above MaxEnergy to MaxEnergy", func() {
			energy := float32(150.0)
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(MaxEnergy))
		})

		It("does not clamp energy within valid range", func() {
			energy := float32(50.0)
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(energy))
		})

		It("does not clamp energy at exactly zero", func() {
			energy := float32(0.0)
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(float32(0.0)))
		})

		It("does not clamp energy at exactly MaxEnergy", func() {
			energy := MaxEnergy
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(MaxEnergy))
		})

		It("handles very large negative values", func() {
			energy := float32(-1000.0)
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(float32(0.0)))
		})

		It("handles very large positive values", func() {
			energy := float32(10000.0)
			clamped := ClampEnergy(energy)
			Expect(clamped).To(Equal(MaxEnergy))
		})
	})

	Describe("Energy Economy Integration", func() {
		It("correctly drains then restores energy", func() {
			initialEnergy := float32(100.0)
			// Drain 3 times
			energy1 := DrainEnergyOnThrust(initialEnergy, true)
			energy2 := DrainEnergyOnThrust(energy1, true)
			energy3 := DrainEnergyOnThrust(energy2, true)
			// Restore once
			finalEnergy := RestoreEnergyOnPickup(energy3)
			expectedEnergy := initialEnergy - 3*ThrustDrainRate + PalletRestoreAmount
			// Clamp to valid range
			if expectedEnergy > MaxEnergy {
				expectedEnergy = MaxEnergy
			}
			Expect(finalEnergy).To(BeNumerically("~", expectedEnergy, 0.001))
		})

		It("correctly restores then drains energy", func() {
			initialEnergy := float32(50.0)
			// Restore once
			energy1 := RestoreEnergyOnPickup(initialEnergy)
			// Drain 2 times
			energy2 := DrainEnergyOnThrust(energy1, true)
			finalEnergy := DrainEnergyOnThrust(energy2, true)
			expectedEnergy := initialEnergy + PalletRestoreAmount - 2*ThrustDrainRate
			Expect(finalEnergy).To(BeNumerically("~", expectedEnergy, 0.001))
		})

		It("handles complex sequence of operations", func() {
			energy := float32(100.0)
			// Drain, drain, restore, drain, restore, restore
			energy = DrainEnergyOnThrust(energy, true)
			energy = DrainEnergyOnThrust(energy, true)
			energy = RestoreEnergyOnPickup(energy)
			energy = DrainEnergyOnThrust(energy, true)
			energy = RestoreEnergyOnPickup(energy)
			energy = RestoreEnergyOnPickup(energy)
			// Should be at max (100 - 0.5*3 + 25*3 = 100 - 1.5 + 75 = 173.5, clamped to 100)
			Expect(energy).To(Equal(MaxEnergy))
		})

		It("maintains energy conservation across operations", func() {
			initialEnergy := float32(100.0)
			energy := initialEnergy
			// Perform multiple operations
			energy = DrainEnergyOnThrust(energy, true)
			energy = DrainEnergyOnThrust(energy, true)
			energy = RestoreEnergyOnPickup(energy)
			energy = DrainEnergyOnThrust(energy, false) // No drain
			energy = RestoreEnergyOnPickup(energy)
			// Verify energy is within valid bounds
			Expect(energy).To(BeNumerically(">=", 0.0))
			Expect(energy).To(BeNumerically("<=", MaxEnergy))
		})

		It("handles draining to zero then restoring", func() {
			// Start with energy equal to drain rate
			energy := ThrustDrainRate
			// Drain to zero
			energy = DrainEnergyOnThrust(energy, true)
			Expect(energy).To(Equal(float32(0.0)))
			// Restore
			energy = RestoreEnergyOnPickup(energy)
			Expect(energy).To(Equal(PalletRestoreAmount))
		})

		It("handles restoring to max then draining", func() {
			// Start near max
			energy := MaxEnergy - PalletRestoreAmount + 1.0
			// Restore to max
			energy = RestoreEnergyOnPickup(energy)
			Expect(energy).To(Equal(MaxEnergy))
			// Drain
			energy = DrainEnergyOnThrust(energy, true)
			Expect(energy).To(BeNumerically("~", MaxEnergy-ThrustDrainRate, 0.001))
		})
	})
})

