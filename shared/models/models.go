package models

import (
	"time"
)

// Patient model
type Patient struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	BirthDate *time.Time `json:"birthDate"`
	Details   string     `json:"details"`

	// One-to-many relationship: a patient can have multiple examinations
	Examinations []Examination `json:"examinations,omitempty"`
}

// Examination model
type Examination struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	PatientID uint       `json:"patientId"` // foreign key for Patient
	ExamDate  *time.Time `json:"examDate" gorm:"not null"`
	Anamnesis string     `json:"anamnesis"`
	Diagnosis string     `json:"diagnosis"`

	// Belongs to
	Patient Patient `json:"patient,omitempty"`

	// One-to-many relationships
	Samples       []Sample       `json:"samples,omitempty"`
	Prescriptions []Prescription `json:"prescriptions,omitempty"`
	Referrals     []Referral     `json:"referrals,omitempty"`
}

// Sample model
type Sample struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ExaminationID uint   `json:"examinationId"` // foreign key for Examination
	SampleType    string `json:"sampleType"`
	Result        string `json:"result"`

	// Belongs to
	Examination Examination `json:"examination,omitempty"`
}

// Prescription model
type Prescription struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ExaminationID uint   `json:"examinationId"` // foreign key for Examination
	Medication    string `json:"medication"`
	Dosage        string `json:"dosage"`
	Instructions  string `json:"instructions"`
	Validated     bool   `json:"validated"`
	Sent          bool   `json:"sent"`

	// Belongs to
	Examination Examination `json:"examination,omitempty"`
}

// Referral model
type Referral struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ExaminationID uint   `json:"examinationId"` // foreign key for Examination
	Specialist    string `json:"specialist"`
	Reason        string `json:"reason"`

	// Belongs to
	Examination Examination `json:"examination,omitempty"`
}
