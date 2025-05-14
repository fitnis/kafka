package services

import (
	"errors"

	"github.com/fitnis/shared/models"
	"gorm.io/gorm"
)

// ReferralService handles database operations for referrals.
type ReferralService struct {
	DB *gorm.DB
}

// NewReferralService creates a new ReferralService.
func NewReferralService(db *gorm.DB) *ReferralService {
	return &ReferralService{DB: db}
}

// CreateReferral adds a new referral to the database.
func (s *ReferralService) CreateReferral(examinationID uint, specialist, reason string) (models.Referral, error) {
	referral := models.Referral{
		ExaminationID: examinationID,
		Specialist:    specialist,
		Reason:        reason,
	}
	result := s.DB.Create(&referral)
	return referral, result.Error
}

// GetReferrals retrieves all referrals from the database.
func (s *ReferralService) GetReferrals() ([]models.Referral, error) {
	var referrals []models.Referral
	result := s.DB.Find(&referrals)
	return referrals, result.Error
}

// GetReferralByID retrieves a referral by its ID.
func (s *ReferralService) GetReferralByID(id uint) (models.Referral, error) {
	var referral models.Referral
	result := s.DB.First(&referral, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Referral{}, errors.New("referral not found")
		}
		return models.Referral{}, result.Error
	}
	return referral, nil
}

// GetReferralsByExaminationID retrieves all referrals for a specific examination.
func (s *ReferralService) GetReferralsByExaminationID(examinationID uint) ([]models.Referral, error) {
	var examReferrals []models.Referral
	result := s.DB.Where("examination_id = ?", examinationID).Find(&examReferrals)
	return examReferrals, result.Error
}

// UpdateReferral updates an existing referral's details.
func (s *ReferralService) UpdateReferral(id uint, specialist, reason string) (models.Referral, error) {
	referral, err := s.GetReferralByID(id)
	if err != nil {
		return models.Referral{}, err
	}

	// Update fields if provided
	if specialist != "" {
		referral.Specialist = specialist
	}
	// Allow clearing reason
	referral.Reason = reason

	result := s.DB.Save(&referral)
	return referral, result.Error
}

// DeleteReferral removes a referral from the database.
func (s *ReferralService) DeleteReferral(id uint) error {
	result := s.DB.Delete(&models.Referral{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("referral not found or already deleted")
	}
	return nil
}
