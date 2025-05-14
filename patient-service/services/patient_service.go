package services

import (
	"errors"
	"time"

	"github.com/fitnis/shared/models"
	"gorm.io/gorm"
)

// PatientService handles database operations for patients.
type PatientService struct {
	DB *gorm.DB
}

// NewPatientService creates a new PatientService.
func NewPatientService(db *gorm.DB) *PatientService {
	return &PatientService{DB: db}
}

// CreatePatient adds a new patient to the database.
func (s *PatientService) CreatePatient(firstName, lastName, details string, birthDate *time.Time) (models.Patient, error) {
	patient := models.Patient{
		FirstName: firstName,
		LastName:  lastName,
		BirthDate: birthDate,
		Details:   details,
	}
	result := s.DB.Create(&patient)
	return patient, result.Error
}

// GetPatients retrieves all patients from the database.
func (s *PatientService) GetPatients() ([]models.Patient, error) {
	var patients []models.Patient
	result := s.DB.Find(&patients)
	return patients, result.Error
}

// GetPatientByID retrieves a patient by their ID.
func (s *PatientService) GetPatientByID(id uint) (models.Patient, error) {
	var patient models.Patient
	result := s.DB.First(&patient, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Patient{}, errors.New("patient not found")
		}
		return models.Patient{}, result.Error
	}
	return patient, nil
}

// UpdatePatient updates an existing patient's details.
func (s *PatientService) UpdatePatient(id uint, firstName, lastName, details string, birthDate time.Time) (models.Patient, error) {
	patient, err := s.GetPatientByID(id)
	if err != nil {
		return models.Patient{}, err
	}

	// Only update fields if they are provided in the request (check non-zero/non-empty)
	// Note: A more robust way might involve using a map[string]interface{} or separate update DTOs
	if firstName != "" {
		patient.FirstName = firstName
	}
	if lastName != "" {
		patient.LastName = lastName
	}
	if !birthDate.IsZero() {
		patient.BirthDate = &birthDate
	}
	// Allow clearing details
	patient.Details = details

	result := s.DB.Save(&patient)
	return patient, result.Error
}

// DeletePatient removes a patient from the database.
func (s *PatientService) DeletePatient(id uint) error {
	result := s.DB.Delete(&models.Patient{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("patient not found or already deleted")
	}
	return nil
}
