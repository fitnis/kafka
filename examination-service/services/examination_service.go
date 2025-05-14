package services

import (
	"errors"
	"time"

	"github.com/fitnis/shared/models"
	"gorm.io/gorm"
)

// ExaminationService handles database operations for examinations.
type ExaminationService struct {
	DB *gorm.DB
}

// NewExaminationService creates a new ExaminationService.
func NewExaminationService(db *gorm.DB) *ExaminationService {
	return &ExaminationService{DB: db}
}

// CreateExamination adds a new examination to the database.
func (s *ExaminationService) CreateExamination(patientID uint, examDate *time.Time, anamnesis, diagnosis string) (models.Examination, error) {
	exam := models.Examination{
		PatientID: patientID,
		ExamDate:  examDate,
		Anamnesis: anamnesis,
		Diagnosis: diagnosis,
	}
	result := s.DB.Create(&exam)
	return exam, result.Error
}

// GetExaminations retrieves all examinations from the database.
func (s *ExaminationService) GetExaminations() ([]models.Examination, error) {
	var examinations []models.Examination
	// Preload associated data if needed, e.g., Patient
	result := s.DB.Preload("Patient").Find(&examinations)
	return examinations, result.Error
}

// GetExaminationByID retrieves an examination by its ID.
func (s *ExaminationService) GetExaminationByID(id uint) (models.Examination, error) {
	var exam models.Examination
	// Preload associated data if needed
	// result := s.DB.Preload("Patient").Preload("Samples").Preload("Prescriptions").Preload("Referrals").First(&exam, id)
	result := s.DB.First(&exam, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Examination{}, errors.New("examination not found")
		}
		return models.Examination{}, result.Error
	}
	return exam, nil
}

// GetExaminationsByPatientID retrieves all examinations for a specific patient.
func (s *ExaminationService) GetExaminationsByPatientID(patientID uint) ([]models.Examination, error) {
	var patientExams []models.Examination
	result := s.DB.Where("patient_id = ?", patientID).Find(&patientExams)
	return patientExams, result.Error
}

// UpdateExamination updates an existing examination's details.
func (s *ExaminationService) UpdateExamination(id uint, examDate time.Time, anamnesis, diagnosis string) (models.Examination, error) {
	exam, err := s.GetExaminationByID(id)
	if err != nil {
		return models.Examination{}, err
	}

	// Update fields if provided
	if !examDate.IsZero() {
		exam.ExamDate = &examDate
	}
	// Allow clearing fields by providing empty strings
	exam.Anamnesis = anamnesis
	exam.Diagnosis = diagnosis

	result := s.DB.Save(&exam)
	return exam, result.Error
}

// DeleteExamination removes an examination from the database.
// Note: Consider cascading deletes or handling related records (samples, prescriptions, referrals)
func (s *ExaminationService) DeleteExamination(id uint) error {
	result := s.DB.Delete(&models.Examination{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("examination not found or already deleted")
	}
	return nil
}
