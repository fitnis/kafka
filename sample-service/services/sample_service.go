package services

import (
	"errors"
	"fmt"

	"github.com/fitnis/prescription-service/services"
	"github.com/fitnis/shared/models"
	"gorm.io/gorm"
)

// SampleService handles database operations for samples.
type SampleService struct {
	DB                  *gorm.DB
	PrescriptionService *services.PrescriptionService // Inject PrescriptionService
}

// NewSampleService creates a new SampleService.
func NewSampleService(db *gorm.DB, ps *services.PrescriptionService) *SampleService {
	return &SampleService{DB: db, PrescriptionService: ps}
}

// CreateSample creates a sample, triggers evaluation, and generates a prescription.
func (s *SampleService) CreateSample(examinationID uint, sampleType, result string) (models.Sample, error) {
	// Create the sample
	sample := models.Sample{
		ExaminationID: examinationID,
		SampleType:    sampleType,
		Result:        result, // Initial result if provided
	}

	// Use transaction to ensure atomicity
	tx := s.DB.Begin()
	if tx.Error != nil {
		return models.Sample{}, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Save sample within transaction
	if err := tx.Create(&sample).Error; err != nil {
		tx.Rollback()
		return models.Sample{}, fmt.Errorf("failed to create sample: %w", err)
	}

	// Automatically evaluate the sample (mock logic)
	evaluationResult := s.evaluateSample(sample)

	// Update sample result within transaction
	sample.Result = evaluationResult
	if err := tx.Save(&sample).Error; err != nil {
		tx.Rollback()
		return models.Sample{}, fmt.Errorf("failed to update sample result: %w", err)
	}

	// Generate prescription based on evaluation (using PrescriptionService within transaction)
	_, err := s.generateAutomaticPrescription(tx, examinationID, evaluationResult)
	if err != nil {
		tx.Rollback()
		return models.Sample{}, fmt.Errorf("failed to generate prescription: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return models.Sample{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sample, nil
}

// GetSamples retrieves all samples.
func (s *SampleService) GetSamples() ([]models.Sample, error) {
	var samples []models.Sample
	result := s.DB.Find(&samples)
	return samples, result.Error
}

// GetSampleByID retrieves a sample by ID.
func (s *SampleService) GetSampleByID(id uint) (models.Sample, error) {
	var sample models.Sample
	result := s.DB.First(&sample, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Sample{}, errors.New("sample not found")
		}
		return models.Sample{}, result.Error
	}
	return sample, nil
}

// GetSamplesByExaminationID retrieves samples for a specific examination.
func (s *SampleService) GetSamplesByExaminationID(examinationID uint) ([]models.Sample, error) {
	var samples []models.Sample
	result := s.DB.Where("examination_id = ?", examinationID).Find(&samples)
	return samples, result.Error
}

// UpdateSample updates an existing sample.
func (s *SampleService) UpdateSample(id uint, sampleType, result string) (models.Sample, error) {
	sample, err := s.GetSampleByID(id)
	if err != nil {
		return models.Sample{}, err
	}

	if sampleType != "" {
		sample.SampleType = sampleType
	}
	// Allow updating/clearing result
	sample.Result = result

	dbResult := s.DB.Save(&sample)
	return sample, dbResult.Error
}

// DeleteSample removes a sample from the database.
func (s *SampleService) DeleteSample(id uint) error {
	result := s.DB.Delete(&models.Sample{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("sample not found or already deleted")
	}
	return nil
}

// evaluateSample (private helper) evaluates a sample and returns analysis results.
func (s *SampleService) evaluateSample(sample models.Sample) string {
	// Mock evaluation process
	var result string
	switch sample.SampleType {
	case "blood":
		result = "Blood sample analysis complete. Parameters within normal range."
	case "urine":
		result = "Urine sample analysis complete. No abnormalities detected."
	case "tissue":
		result = "Tissue sample analysis complete. No pathological changes observed."
	default:
		result = fmt.Sprintf("%s sample analyzed. Results appear normal.", sample.SampleType)
	}
	return result
}

// generateAutomaticPrescription (private helper) creates a prescription based on evaluation.
// It now accepts a *gorm.DB (potentially a transaction) to ensure atomicity.
func (s *SampleService) generateAutomaticPrescription(db *gorm.DB, examinationID uint, evaluationResult string) (models.Prescription, error) {
	// Mock prescription generation logic
	var medication, dosage, instructions string
	if evaluationResult == "" || len(evaluationResult) < 10 {
		medication = "Generic medication"
		dosage = "Standard dosage"
		instructions = "Take as directed"
	} else if evaluationResult == "Blood sample analysis complete. Parameters within normal range." {
		medication = "Iron supplement"
		dosage = "One tablet daily"
		instructions = "Take with food"
	} else if evaluationResult == "Urine sample analysis complete. No abnormalities detected." {
		medication = "Vitamin C"
		dosage = "500mg daily"
		instructions = "Take with water"
	} else {
		medication = "General antibiotic"
		dosage = "One pill twice daily"
		instructions = "Take for 7 days"
	}

	// Use the injected PrescriptionService, passing the transaction DB
	tempPrescriptionService := services.NewPrescriptionService(db) // Use the transaction DB
	return tempPrescriptionService.CreatePrescription(examinationID, medication, dosage, instructions)
}
