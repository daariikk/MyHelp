package domain

import (
	"testing"
	"time"
)

func TestNewSpecialization(t *testing.T) {
	tests := []struct {
		name                 string
		id                   int
		specialization       string
		specializationDoctor string
		description          string
		expected             Specialization
	}{
		{
			name:                 "All fields filled",
			id:                   1,
			specialization:       "Cardiology",
			specializationDoctor: "Cardiologist",
			description:          "Heart related treatments",
			expected: Specialization{
				ID:                   1,
				Specialization:       "Cardiology",
				SpecializationDoctor: "Cardiologist",
				Description:          "Heart related treatments",
			},
		},
		{
			name:                 "Empty description",
			id:                   2,
			specialization:       "Neurology",
			specializationDoctor: "Neurologist",
			description:          "",
			expected: Specialization{
				ID:                   2,
				Specialization:       "Neurology",
				SpecializationDoctor: "Neurologist",
				Description:          "",
			},
		},
		{
			name:                 "Zero ID",
			id:                   0,
			specialization:       "General",
			specializationDoctor: "General Practitioner",
			description:          "General health",
			expected: Specialization{
				ID:                   0,
				Specialization:       "General",
				SpecializationDoctor: "General Practitioner",
				Description:          "General health",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSpecialization(
				tt.id,
				tt.specialization,
				tt.specializationDoctor,
				tt.description,
			)

			if result != tt.expected {
				t.Errorf("NewSpecialization() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSpecializationFields(t *testing.T) {
	spec := NewSpecialization(
		1,
		"Pediatrics",
		"Pediatrician",
		"Child healthcare",
	)

	if spec.ID != 1 {
		t.Errorf("Expected ID 1, got %d", spec.ID)
	}

	if spec.Specialization != "Pediatrics" {
		t.Errorf("Expected Specialization 'Pediatrics', got '%s'", spec.Specialization)
	}

	if spec.SpecializationDoctor != "Pediatrician" {
		t.Errorf("Expected SpecializationDoctor 'Pediatrician', got '%s'", spec.SpecializationDoctor)
	}

	if spec.Description != "Child healthcare" {
		t.Errorf("Expected Description 'Child healthcare', got '%s'", spec.Description)
	}
}

func TestNewDoctor(t *testing.T) {
	tests := []struct {
		name           string
		input          Doctor
		expected       Doctor
		expectedRating float64
	}{
		{
			name: "All fields filled",
			input: NewDoctor(
				1,
				"Ivanov",
				"Ivan",
				"Ivanovich",
				"Cardiology",
				"First Medical University",
				"Head of Department",
				4.8,
				"/photos/ivanov.jpg",
			),
			expected: Doctor{
				Id:             1,
				Surname:        "Ivanov",
				Name:           "Ivan",
				Patronymic:     "Ivanovich",
				Specialization: "Cardiology",
				Education:      "First Medical University",
				Progress:       "Head of Department",
				Rating:         4.8,
				PhotoPath:      "/photos/ivanov.jpg",
			},
			expectedRating: 4.8,
		},
		{
			name: "Empty optional fields",
			input: NewDoctor(
				2,
				"Petrova",
				"Maria",
				"",
				"Pediatrics",
				"Second Medical College",
				"",
				4.5,
				"",
			),
			expected: Doctor{
				Id:             2,
				Surname:        "Petrova",
				Name:           "Maria",
				Patronymic:     "",
				Specialization: "Pediatrics",
				Education:      "Second Medical College",
				Progress:       "",
				Rating:         4.5,
				PhotoPath:      "",
			},
			expectedRating: 4.5,
		},
		{
			name: "Zero rating",
			input: NewDoctor(
				3,
				"Sidorov",
				"Alexey",
				"Viktorovich",
				"Neurology",
				"Third Medical Institute",
				"Researcher",
				0,
				"/photos/sidorov.jpg",
			),
			expected: Doctor{
				Id:             3,
				Surname:        "Sidorov",
				Name:           "Alexey",
				Patronymic:     "Viktorovich",
				Specialization: "Neurology",
				Education:      "Third Medical Institute",
				Progress:       "Researcher",
				Rating:         0,
				PhotoPath:      "/photos/sidorov.jpg",
			},
			expectedRating: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != tt.expected {
				t.Errorf("NewDoctor() = %+v, want %+v", tt.input, tt.expected)
			}

			if tt.input.Rating != tt.expectedRating {
				t.Errorf("Expected rating %.1f, got %.1f", tt.expectedRating, tt.input.Rating)
			}
		})
	}
}

func TestDoctorFields(t *testing.T) {
	doctor := NewDoctor(
		10,
		"Kuznetsova",
		"Olga",
		"Sergeevna",
		"Dermatology",
		"Medical Academy",
		"Senior Doctor",
		4.9,
		"/photos/kuznetsova.jpg",
	)

	if doctor.Id != 10 {
		t.Errorf("Expected ID 10, got %d", doctor.Id)
	}

	if doctor.Surname != "Kuznetsova" {
		t.Errorf("Expected Surname 'Kuznetsova', got '%s'", doctor.Surname)
	}

	if doctor.Name != "Olga" {
		t.Errorf("Expected Name 'Olga', got '%s'", doctor.Name)
	}

	if doctor.Patronymic != "Sergeevna" {
		t.Errorf("Expected Patronymic 'Sergeevna', got '%s'", doctor.Patronymic)
	}

	if doctor.Specialization != "Dermatology" {
		t.Errorf("Expected Specialization 'Dermatology', got '%s'", doctor.Specialization)
	}

	if doctor.Rating != 4.9 {
		t.Errorf("Expected Rating 4.9, got %.1f", doctor.Rating)
	}
}

func TestNewRecord(t *testing.T) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	end := start.Add(time.Hour * 1)

	tests := []struct {
		name         string
		input        Record
		expected     Record
		expectedTime time.Time
	}{
		{
			name: "Available record",
			input: NewRecord(
				1,
				101,
				now,
				start,
				end,
				true,
			),
			expected: Record{
				ID:          1,
				DoctorId:    101,
				Date:        now,
				Start:       start,
				End:         end,
				IsAvailable: true,
			},
			expectedTime: start,
		},
		{
			name: "Unavailable record",
			input: NewRecord(
				2,
				102,
				now.AddDate(0, 0, 1),
				start.Add(time.Hour*2),
				end.Add(time.Hour*2),
				false,
			),
			expected: Record{
				ID:          2,
				DoctorId:    102,
				Date:        now.AddDate(0, 0, 1),
				Start:       start.Add(time.Hour * 2),
				End:         end.Add(time.Hour * 2),
				IsAvailable: false,
			},
			expectedTime: start.Add(time.Hour * 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.ID != tt.expected.ID {
				t.Errorf("Expected ID %d, got %d", tt.expected.ID, tt.input.ID)
			}
			if tt.input.DoctorId != tt.expected.DoctorId {
				t.Errorf("Expected DoctorId %d, got %d", tt.expected.DoctorId, tt.input.DoctorId)
			}
			if !tt.input.Date.Equal(tt.expected.Date) {
				t.Errorf("Expected Date %v, got %v", tt.expected.Date, tt.input.Date)
			}
			if !tt.input.Start.Equal(tt.expectedTime) {
				t.Errorf("Expected Start time %v, got %v", tt.expectedTime, tt.input.Start)
			}
			if tt.input.IsAvailable != tt.expected.IsAvailable {
				t.Errorf("Expected IsAvailable %v, got %v", tt.expected.IsAvailable, tt.input.IsAvailable)
			}
		})
	}
}

func TestNewSchedule(t *testing.T) {
	now := time.Now()
	records := []Record{
		NewRecord(1, 101, now, now.Add(time.Hour*9), now.Add(time.Hour*10), true),
		NewRecord(2, 101, now, now.Add(time.Hour*10), now.Add(time.Hour*11), false),
	}

	schedule := NewSchedule(records)

	if len(schedule.Records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(schedule.Records))
	}

	if schedule.Records[0].DoctorId != 101 {
		t.Errorf("First record should have DoctorId 101, got %d", schedule.Records[0].DoctorId)
	}

	if schedule.Records[1].IsAvailable {
		t.Error("Second record should be unavailable")
	}
}

func TestNewScheduleInfoDTO(t *testing.T) {
	doctor := NewDoctor(
		101,
		"Ivanov",
		"Ivan",
		"Ivanovich",
		"Cardiology",
		"First Medical",
		"Head Doctor",
		4.8,
		"/photos/ivanov.jpg",
	)

	now := time.Now()
	records := []Record{
		NewRecord(1, 101, now, now.Add(time.Hour*9), now.Add(time.Hour*10), true),
	}
	schedule := NewSchedule(records)

	dto := NewScheduleInfoDTO(doctor, schedule)

	if dto.Doctor.Id != 101 {
		t.Errorf("Expected Doctor ID 101, got %d", dto.Doctor.Id)
	}

	if len(dto.Schedule.Records) != 1 {
		t.Errorf("Expected 1 schedule record, got %d", len(dto.Schedule.Records))
	}

	if dto.Schedule.Records[0].ID != 1 {
		t.Errorf("Expected record ID 1, got %d", dto.Schedule.Records[0].ID)
	}
}
