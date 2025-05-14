package services

import (
	"errors"

	"github.com/fitnis/shared/models"
	"gorm.io/gorm"
)

// PrescriptionService handles database operations for prescriptions.
type PrescriptionService struct {
	DB *gorm.DB
}

// NewPrescriptionService creates a new PrescriptionService.
// It accepts a *gorm.DB, which could be the main DB or a transaction DB.
func NewPrescriptionService(db *gorm.DB) *PrescriptionService {
	return &PrescriptionService{DB: db}
}

// CreatePrescription adds a new prescription to the database.
func (s *PrescriptionService) CreatePrescription(examinationID uint, medication, dosage, instructions string) (models.Prescription, error) {
	prescription := models.Prescription{
		ExaminationID: examinationID,
		Medication:    medication,
		Dosage:        dosage,
		Instructions:  instructions,
		Validated:     false, // Default values
		Sent:          false,
	}
	// Use the DB instance associated with the service (could be main DB or transaction)
	result := s.DB.Create(&prescription)
	return prescription, result.Error
}

// GetPrescriptions retrieves all prescriptions from the database.
func (s *PrescriptionService) GetPrescriptions() ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	result := s.DB.Find(&prescriptions)
	return prescriptions, result.Error
}

// GetPrescriptionByID retrieves a prescription by its ID.
func (s *PrescriptionService) GetPrescriptionByID(id uint) (models.Prescription, error) {
	var prescription models.Prescription
	result := s.DB.First(&prescription, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Prescription{}, errors.New("prescription not found")
		}
		return models.Prescription{}, result.Error
	}
	return prescription, nil
}

// GetPrescriptionsByExaminationID retrieves all prescriptions for a specific examination.
func (s *PrescriptionService) GetPrescriptionsByExaminationID(examinationID uint) ([]models.Prescription, error) {
	var examPrescriptions []models.Prescription
	result := s.DB.Where("examination_id = ?", examinationID).Find(&examPrescriptions)
	return examPrescriptions, result.Error
}

// UpdatePrescription updates an existing prescription's details.
func (s *PrescriptionService) UpdatePrescription(id uint, medication, dosage, instructions string, validated, sent *bool) (models.Prescription, error) {
	prescription, err := s.GetPrescriptionByID(id)
	if err != nil {
		return models.Prescription{}, err
	}

	// Update fields if provided (use pointers for bools to distinguish between false and not provided)
	if medication != "" {
		prescription.Medication = medication
	}
	if dosage != "" {
		prescription.Dosage = dosage
	}
	// Allow clearing instructions
	prescription.Instructions = instructions

	if validated != nil {
		prescription.Validated = *validated
	}
	if sent != nil {
		// Business rule: Cannot mark as sent if not validated
		if *sent && !prescription.Validated {
			return models.Prescription{}, errors.New("cannot mark prescription as sent before it is validated")
		}
		prescription.Sent = *sent
	}

	result := s.DB.Save(&prescription)
	return prescription, result.Error
}

// ValidatePrescription marks a prescription as validated.
func (s *PrescriptionService) ValidatePrescription(id uint) (models.Prescription, error) {
	prescription, err := s.GetPrescriptionByID(id)
	if err != nil {
		return models.Prescription{}, err
	}

	if prescription.Validated {
		// Optionally return an error or just the current state if already validated
		return prescription, nil // Or: errors.New("prescription already validated")
	}

	prescription.Validated = true
	result := s.DB.Save(&prescription)
	return prescription, result.Error
}

// SendPrescription marks a validated prescription as sent.
func (s *PrescriptionService) SendPrescription(id uint) (models.Prescription, error) {
	prescription, err := s.GetPrescriptionByID(id)
	if err != nil {
		return models.Prescription{}, err
	}

	if !prescription.Validated {
		return models.Prescription{}, errors.New("prescription must be validated before sending")
	}

	if prescription.Sent {
		// Optionally return an error or just the current state if already sent
		return prescription, nil // Or: errors.New("prescription already sent")
	}

	prescription.Sent = true
	result := s.DB.Save(&prescription)
	return prescription, result.Error
}

// DeletePrescription removes a prescription from the database.
func (s *PrescriptionService) DeletePrescription(id uint) error {
	result := s.DB.Delete(&models.Prescription{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("prescription not found or already deleted")
	}
	return nil
}
