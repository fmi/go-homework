package main

import (
	"encoding/json"
	"fmt"
)

type Susi struct {
	courses     map[string]*Course
	students    map[int]*Student
	enrollments map[int]map[string]*Enrollment // [fn][identifier]
}

func NewSusi() *Susi {
	s := new(Susi)

	s.courses = make(map[string]*Course)
	s.students = make(map[int]*Student)
	s.enrollments = make(map[int]map[string]*Enrollment)

	return s
}

type SusiError struct {
	Student *Student
	Course  *Course
	message string
}

func (se *SusiError) Error() string {
	return se.message
}

func (s *Susi) newError(message string, student *Student, course *Course) *SusiError {
	se := new(SusiError)
	se.message = message
	se.Student = student
	se.Course = course
	return se
}

func (s *Susi) studentNotFoundError(fn int) string {
	return fmt.Sprintf("Няма студент с факултетен номер %d!", fn)
}

func (s *Susi) studentAlreadyExistsError(fn int) string {
	return fmt.Sprintf("Студент с факултетен номер %d вече съществува!", fn)
}

type studentRequest struct {
	Student Student `json:"student"`
}

type Student struct {
	FacultyNumber int    `json:"faculty_number"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Master        bool   `json:"Master"`
	AcademicYear  int    `json:"academic_year"`
}

func (s *Student) String() string {
	return fmt.Sprintf("%d %s %s", s.FacultyNumber, s.FirstName, s.LastName)
}

func (s *Susi) studentShouldExist(fn int) error {
	if _, ok := s.students[fn]; !ok {
		return s.newError(s.studentNotFoundError(fn), nil, nil)
	}

	return nil
}

func (s *Susi) studentShouldNotExist(fn int) error {
	if existing, ok := s.students[fn]; ok {
		return s.newError(s.studentAlreadyExistsError(fn), existing, nil)
	}

	return nil
}

func (s *Susi) AddStudent(request []byte) error {
	var student *Student
	err := json.Unmarshal(request, &student)

	if err != nil {
		return s.newError(err.Error(), nil, nil)
	}

	if err := s.studentShouldNotExist(student.FacultyNumber); err != nil {
		return err
	}

	s.students[student.FacultyNumber] = student

	return nil
}

func (s *Susi) FindStudent(fn int) (*Student, error) {
	if student, ok := s.students[fn]; ok {
		return student, nil
	}

	return nil, s.newError(s.studentNotFoundError(fn), nil, nil)
}

func (s *Susi) courseNotFoundError(identifier string) string {
	return fmt.Sprintf("Няма курс с identifier - %s!", identifier)
}

func (s *Susi) courseAlreadyExistsError(identifier string) string {
	return fmt.Sprintf("Курс с identifier %s вече съществува!", identifier)
}

func (s *Susi) courseIsFullError(identifier string) string {
	return fmt.Sprintf("Няма свободни места за курс с identifier - %s!", identifier)
}

type courseRequest struct {
	Course Course `json:"course"`
}

type Course struct {
	CourseName          string `json:"course_name"`
	CourseIdentifier    string `json:"course_identifier"`
	MinimumAcademicYear int    `json:"minimum_academic_year"`
	MastersOnly         bool   `json:"Masters_only"`
	AvailablePlaces     int    `json:"available_places"`
}

func (c *Course) String() string {
	return fmt.Sprintf("%s %s", c.CourseIdentifier, c.CourseName)
}

func (s *Susi) courseShouldExist(identifier string) error {
	if _, ok := s.courses[identifier]; !ok {
		return s.newError(s.courseNotFoundError(identifier), nil, nil)
	}

	return nil
}

func (s *Susi) courseShouldNotExist(identifier string) error {
	if existing, ok := s.courses[identifier]; ok {
		return s.newError(s.courseAlreadyExistsError(identifier), nil, existing)
	}

	return nil
}

func (s *Susi) AddCourse(request []byte) error {
	var course *Course
	err := json.Unmarshal(request, &course)

	if err != nil {
		return s.newError(err.Error(), nil, nil)
	}

	if err := s.courseShouldNotExist(course.CourseIdentifier); err != nil {
		return err
	}

	s.courses[course.CourseIdentifier] = course

	return nil
}

func (s *Susi) FindCourse(identifier string) (*Course, error) {
	if course, ok := s.courses[identifier]; ok {
		return course, nil
	}

	return nil, s.newError(s.courseNotFoundError(identifier), nil, nil)
}

func (s *Susi) enrollmentAlreadyExistsError(fn int, identifier string) string {
	return fmt.Sprintf("Студент с факултетен номер %d вече е записан за курс с identifier %s!", fn, identifier)
}

func (s *Susi) enrollmentNotFoundError(fn int, identifier string) string {
	return fmt.Sprintf("Студент с факултетен номер %d не е записан за курса с identifier %s!", fn, identifier)
}

type Enrollment struct {
	Student *Student
	Course  *Course
}

type enrollmentRequest struct {
	FacultyNumber    int    `json:"faculty_number"`
	CourseIdentifier string `json:"course_identifier"`
}

func (s *Susi) studentCannotEnrollError(student *Student, course *Course) string {
	return fmt.Sprintf("%s %s не покрива изискванията за %s!", student.FirstName, student.LastName, course.CourseName)
}

func (s *Susi) studentShouldBeAbleToEnroll(student *Student, course *Course) error {
	if student.Master {
		return nil
	}

	if course.MastersOnly || course.MinimumAcademicYear > student.AcademicYear {
		return s.newError(s.studentCannotEnrollError(student, course), student, course)
	}

	return nil
}

func (s *Susi) Enroll(request []byte) error {
	var er *enrollmentRequest

	err := json.Unmarshal(request, &er)

	if err != nil {
		return s.newError(err.Error(), nil, nil)
	}

	if err := s.studentShouldExist(er.FacultyNumber); err != nil {
		return err
	}

	student := s.students[er.FacultyNumber]

	if err := s.courseShouldExist(er.CourseIdentifier); err != nil {
		return err
	}

	course := s.courses[er.CourseIdentifier]

	if studentEnrollments, ok := s.enrollments[student.FacultyNumber]; ok {
		if _, ok := studentEnrollments[course.CourseIdentifier]; ok {
			return s.newError(s.enrollmentAlreadyExistsError(student.FacultyNumber, course.CourseIdentifier), student, course)
		}
	}

	if course.AvailablePlaces <= 0 {
		return s.newError(s.courseIsFullError(er.CourseIdentifier), student, course)
	}

	if err := s.studentShouldBeAbleToEnroll(student, course); err != nil {
		return err
	}

	if _, ok := s.enrollments[student.FacultyNumber]; !ok {
		s.enrollments[student.FacultyNumber] = make(map[string]*Enrollment)
	}

	s.enrollments[student.FacultyNumber][course.CourseIdentifier] = s.newEnrollment(student, course)
	course.AvailablePlaces -= 1

	return nil
}

func (s *Susi) newEnrollment(student *Student, course *Course) *Enrollment {
	e := new(Enrollment)
	e.Student = student
	e.Course = course
	return e
}

func (s *Susi) FindEnrollment(fn int, identifier string) (*Enrollment, error) {
	if err := s.studentShouldExist(fn); err != nil {
		return nil, err
	}

	student := s.students[fn]

	if err := s.courseShouldExist(identifier); err != nil {
		return nil, err
	}

	course := s.courses[identifier]

	if studentEnrollments, ok := s.enrollments[fn]; ok {
		if enrollment, ok := studentEnrollments[identifier]; ok {
			return enrollment, nil
		}
	}

	return nil, s.newError(s.enrollmentNotFoundError(fn, identifier), student, course)
}
