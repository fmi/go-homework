package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type SusiTest struct {
	students map[int][]byte
	courses  map[string][]byte
}

func newSusiTest() (*SusiTest, *Susi) {
	st := new(SusiTest)
	st.students = map[int][]byte{
		11111: []byte(`{"faculty_number":11111,"first_name":"Test","last_name":"One","master":false,"academic_year":1}`),
		11112: []byte(`{"faculty_number":11111,"first_name":"Wrong","last_name":"One","master":false,"academic_year":1}`),
		22222: []byte(`{"faculty_number":22222,"first_name":"Test","last_name":"Two","master":false,"academic_year":2}`),
		33333: []byte(`{"faculty_number":33333,"first_name":"Test","last_name":"Three","master":false,"academic_year":3}`),
		44444: []byte(`{"faculty_number":44444,"first_name":"Test","last_name":"Four","master":false,"academic_year":4}`),
		55555: []byte(`{"faculty_number":55555,"first_name":"Test","last_name":"Master","master":true,"academic_year":0}`),
	}

	st.courses = map[string][]byte{
		"AR":   []byte(`{"course_name":"Advanced Robotics","course_identifier":"AR","minimum_academic_year":3,"masters_only":false,"available_places":2}`),
		"R101": []byte(`{"course_name":"Robotics 101","course_identifier":"R101","minimum_academic_year":1,"masters_only":false,"available_places":2}`),
		"MO":   []byte(`{"course_name":"Masters Only","course_identifier":"MO","minimum_academic_year":0,"masters_only":true,"available_places":2}`),
		"MO2":  []byte(`{"course_name":"Masters Only Wrong","course_identifier":"MO","minimum_academic_year":0,"masters_only":true,"available_places":2}`),
		"FC":   []byte(`{"course_name":"Full Course","course_identifier":"FC","minimum_academic_year":0,"masters_only":false,"available_places":0}`),
	}

	return st, NewSusi()
}

func toInt(ap interface{}) (int, error) {
	var found int

	switch ap.(type) {
	case int:
		found = ap.(int)
	case int64:
		found = int(ap.(int64))
	case int32:
		found = int(ap.(int32))
	case int16:
		found = int(ap.(int16))
	case int8:
		found = int(ap.(int8))
	default:
		return 0, fmt.Errorf("Wrong type")
	}

	return found, nil
}

func (st *SusiTest) AddStudents(s *Susi, fns ...int) error {
	for _, fn := range fns {
		err := s.AddStudent(st.students[fn])
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *SusiTest) AddCourses(s *Susi, identifiers ...string) error {
	for _, identifier := range identifiers {
		err := s.AddCourse(st.courses[identifier])
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *SusiTest) Enroll(s *Susi, fn int, identifier string) error {
	payload := []byte(fmt.Sprintf("{\"faculty_number\":%d,\"course_identifier\":\"%s\"}", fn, identifier))
	return s.Enroll(payload)
}

// Errors
func (st *SusiTest) studentCannotEnrollError(student *Student, course *Course) string {
	return fmt.Sprintf("%s %s не покрива изискванията за %s!", student.FirstName, student.LastName, course.CourseName)
}

func (st *SusiTest) studentNotFoundError(fn int) string {
	return fmt.Sprintf("Няма студент с факултетен номер %d!", fn)
}

func (st *SusiTest) studentAlreadyExistsError(fn int) string {
	return fmt.Sprintf("Студент с факултетен номер %d вече съществува!", fn)
}

func (st *SusiTest) courseNotFoundError(identifier string) string {
	return fmt.Sprintf("Няма курс с identifier - %s!", identifier)
}

func (st *SusiTest) courseAlreadyExistsError(identifier string) string {
	return fmt.Sprintf("Курс с identifier %s вече съществува!", identifier)
}

func (st *SusiTest) courseIsFullError(identifier string) string {
	return fmt.Sprintf("Няма свободни места за курс с identifier - %s!", identifier)
}

func (st *SusiTest) enrollmentAlreadyExistsError(fn int, identifier string) string {
	return fmt.Sprintf("Студент с факултетен номер %d вече е записан за курс с identifier %s!", fn, identifier)
}

func (st *SusiTest) enrollmentNotFoundError(fn int, identifier string) string {
	return fmt.Sprintf("Студент с факултетен номер %d не е записан за курса с identifier %s!", fn, identifier)
}

// Tests

func TestWithBadJson(t *testing.T) {
	var placeholder *Student
	_, s := newSusiTest()
	payload := []byte(`"bad_bad_json"`)
	expected := json.Unmarshal(payload, &placeholder)
	got := s.AddStudent(payload)

	if got == nil {
		t.Error("Expected json error, but didn't get one (recieved nil)")
	}

	if got.Error() != expected.Error() {
		t.Errorf("Expected: %s, got: %s", expected.Error(), got.Error())
	}
}

func TestAddStudent(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}
}

func TestAddTheSameStudentTwice(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 22222, 11111)
	if err == nil {
		t.Error("Expected to recieve an error when adding the same student twice!")
	}

	expected := st.studentAlreadyExistsError(11111)
	got := err.Error()

	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestFindExistingStudent(t *testing.T) {
	st, s := newSusiTest()
	_ = st.AddStudents(s, 22222)
	student, err := s.FindStudent(22222)
	if err != nil {
		t.Errorf("Failed to get a student, recieved: %s!", "error")
	}

	if student.FirstName != "Test" && student.LastName != "Two" {
		t.Errorf("Recieved wrong student: %v", student)
	}
}

func TestFindMissingStudent(t *testing.T) {
	st, s := newSusiTest()
	_, err := s.FindStudent(22222)

	if err == nil {
		t.Error("Expected to recieve an error when getting an missing student!")
	}

	got := err.Error()
	expected := st.studentNotFoundError(22222)
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestAddCourse(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddCourses(s, "AR")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}
}

func TestAddTheSameCourseTwice(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddCourses(s, "AR", "R101", "AR")
	if err == nil {
		t.Error("Expected to recieve an error when adding the same course twice!")
	}

	expected := st.courseAlreadyExistsError("AR")
	got := err.Error()

	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestFindExistingCourse(t *testing.T) {
	st, s := newSusiTest()
	_ = st.AddCourses(s, "R101")
	course, err := s.FindCourse("R101")
	if err != nil {
		t.Errorf("Failed to get a course, recieved: %s!", "error")
	}

	if course.CourseName != "Robotics 101" && course.CourseIdentifier != "R101" {
		t.Errorf("Recieved wrong course: %v", course)
	}
}

func TestFindMissingCourse(t *testing.T) {
	st, s := newSusiTest()
	_, err := s.FindCourse("R101")

	if err == nil {
		t.Error("Expected to recieve an error when getting an missing course!")
	}

	got := err.Error()
	expected := st.courseNotFoundError("R101")
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestEnroll(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 22222)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "AR", "R101", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "R101")

	if err != nil {
		t.Errorf("Failed to enroll in a course, recieved: %s", "error")
	}

	err = st.Enroll(s, 22222, "R101")
	if err != nil {
		t.Errorf("Failed to enroll in a course, recieved: %s", "error")
	}
}

func TestEnrollWithMissingStudent(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "AR", "R101", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 22222, "R101")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling with missing student!")
	}

	got := err.Error()
	expected := st.studentNotFoundError(22222)
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestEnrollWithMissingCourse(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "AR", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "R101")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling in missing course!")
	}

	got := err.Error()
	expected := st.courseNotFoundError("R101")
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestEnrollMoreThanAvailablePlaces(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 22222, 33333)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "R101", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "R101")
	if err != nil {
		t.Errorf("Failed to enroll the first student, got: %s", "error")
	}

	err = st.Enroll(s, 22222, "R101")
	if err != nil {
		t.Errorf("Failed to enroll the second student, got: %s", "error")
	}

	err = st.Enroll(s, 33333, "R101")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling the third student!")
	}

	got := err.Error()
	expected := st.courseIsFullError("R101")
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestEnrollInFullCourse(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 22222, 33333)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "R101", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "FC")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling in full course!")
	}

	got := err.Error()
	expected := st.courseIsFullError("FC")
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestChangeAvailablePlacesOnEnrollment(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 22222, 33333)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "R101", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	course, err := s.FindCourse("R101")
	if err != nil {
		t.Errorf("Failed to get the course, got error")
	}

	var found int

	if found, err = toInt(course.AvailablePlaces); err != nil {
		t.Fatalf("Course.AvailablePlaces: %s. Expected int but was not.", err)
	}

	expectedAvailablePlaces := 2
	if found != expectedAvailablePlaces {
		t.Errorf("Wrong number of available places! Got: %d, expected: %d!", found, expectedAvailablePlaces)
	}

	err = st.Enroll(s, 11111, "R101")
	if err != nil {
		t.Errorf("Failed to enroll the first student, got: %s", "error")
	}

	if found, err = toInt(course.AvailablePlaces); err != nil {
		t.Fatal(err)
	}

	expectedAvailablePlaces -= 1
	if found != expectedAvailablePlaces {
		t.Errorf("Wrong number of available places! Got: %d, expected: %d!", found, expectedAvailablePlaces)
	}

	err = st.Enroll(s, 22222, "R101")
	if err != nil {
		t.Errorf("Failed to enroll the second student, got: %s", "error")
	}

	if found, err = toInt(course.AvailablePlaces); err != nil {
		t.Fatal(err)
	}

	expectedAvailablePlaces -= 1
	if found != expectedAvailablePlaces {
		t.Errorf("Wrong number of available places! Got: %d, expected: %d!", found, expectedAvailablePlaces)
	}
}

func TestEnrollTwiceInTheSameCourseWithTheSameUser(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "R101", "FC")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "R101")
	if err != nil {
		t.Errorf("Failed to enroll the first time, got: %s", "error")
	}

	err = st.Enroll(s, 11111, "R101")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling twise in the same course with the same user!")
	}

	got := err.Error()
	expected := st.enrollmentAlreadyExistsError(11111, "R101")
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestEnrollWhenTheRequirementsAreNotMet(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "R101", "AR")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "AR")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling in a course where the student doesn't meet the requirements!")
	}

	student, _ := s.FindStudent(11111)
	course, _ := s.FindCourse("AR")

	got := err.Error()
	expected := st.studentCannotEnrollError(student, course)
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestEnrollInMasterOnlyCourse(t *testing.T) {
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 55555)
	if err != nil {
		t.Errorf("Failed to add a student, recieved: %s!", "error")
	}

	err = st.AddCourses(s, "MO", "AR")
	if err != nil {
		t.Errorf("Failed to add a course, recieved: %s!", "error")
	}

	err = st.Enroll(s, 55555, "MO")
	if err != nil {
		t.Errorf("Failed to enroll in a master only course when the student is a master, recieved: %s!", "error")
	}

	err = st.Enroll(s, 11111, "MO")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling in a master only course where the student is not a master!")
	}

	student, _ := s.FindStudent(11111)
	course, _ := s.FindCourse("MO")

	got := err.Error()
	expected := st.studentCannotEnrollError(student, course)
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestSusiErrorOnEnrollment(t *testing.T) {
	defer func() {
		if msg := recover(); msg != nil {
			t.Fatalf("didn't bring a towel: %s", msg)
		}
	}()
	st, s := newSusiTest()
	err := st.AddStudents(s, 11111, 55555)
	err = st.AddCourses(s, "MO", "AR")
	err = st.Enroll(s, 11111, "MO")
	if err == nil {
		t.Error("Expected to recieve an error when enrolling in a master only course where the student is not a master!")
	}

	student, _ := s.FindStudent(11111)
	course, _ := s.FindCourse("MO")

	got := err.Error()
	expected := st.studentCannotEnrollError(student, course)
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}

	errorType := reflect.TypeOf(err).String()
	if errorType != "*main.SusiError" && errorType != "*SusiError" {
		t.Errorf("Expected error to be *main.SusiError, but was: %s", errorType)
	}

	susiErr := err.(*SusiError)

	if susiErr.Course.CourseIdentifier != course.CourseIdentifier || susiErr.Course.CourseName != course.CourseName {
		t.Errorf("Expected susiErr.Course to be %V, but was %V", course, susiErr.Course)
	}

	if susiErr.Student.FacultyNumber != student.FacultyNumber || susiErr.Student.FirstName != student.FirstName {
		t.Errorf("Expected susiErr.Student to be %V, but was %V", student, susiErr.Student)
	}
}

func TestFindEnrollment(t *testing.T) {
	st, s := newSusiTest()
	_ = st.AddStudents(s, 11111, 22222)
	_ = st.AddCourses(s, "AR", "R101", "FC")
	_ = st.Enroll(s, 11111, "R101")

	enrollment, err := s.FindEnrollment(11111, "R101")
	if err != nil {
		t.Errorf("Failed to get enrollment, got: %s", "error")
	}
	course, _ := s.FindCourse("R101")
	student, _ := s.FindStudent(11111)

	if enrollment.Course != course {
		t.Errorf("Expected enrollment.Course to be %V, but was %V", course, enrollment.Course)
	}

	if enrollment.Student != student {
		t.Errorf("Expected enrollment.Student to be %V, but was %V", student, enrollment.Student)
	}
}

func TestFindMissingEnrollment(t *testing.T) {
	st, s := newSusiTest()
	_ = st.AddStudents(s, 11111, 22222)
	_ = st.AddCourses(s, "AR", "R101", "FC")
	_ = st.Enroll(s, 11111, "R101")

	enrollment, err := s.FindEnrollment(22222, "R101")
	if err == nil {
		t.Error("Expected to recieve an error when getting an missing enrollment!")
	}

	if enrollment != nil {
		t.Errorf("Expected to get nil enrollment when enrollment is missing, got: %V", enrollment)
	}

	got := err.Error()
	expected := st.enrollmentNotFoundError(22222, "R101")
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

func TestStudentImplementStringer(t *testing.T) {
	st, s := newSusiTest()
	_ = st.AddStudents(s, 11111, 22222)
	student, _ := s.FindStudent(11111)

	got := student.String()
	expected := "11111 Test One"
	if got != expected {
		t.Errorf("Student#String failed! Expected: %s, got: %s", expected, got)
	}
}

func TestCourseImplementStringer(t *testing.T) {
	st, s := newSusiTest()
	_ = st.AddCourses(s, "AR", "R101")
	course, _ := s.FindCourse("AR")

	got := course.String()
	expected := "AR Advanced Robotics"
	if got != expected {
		t.Errorf("Course#String failed! Expected: %s, got: %s", expected, got)
	}
}
