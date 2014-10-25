# Susi
_Ще се записваме за избираеми..._

След като вече сте запознати с типовете в Go и знаете какво е [json](http://json.org/) може да си поиграем на "записване на курсове".

__Напишете функция, която да връща указател към обект от тип Susi:__

    func NewSusi() *Susi


## Типове
### Susi

#### Елементи
Ние няма да се интересуваме как съхранявате елементите в този тип и единствено ще тестваме неговите методи.

#### Методи

    func (s *Susi) AddStudent(request []byte) error
    func (s *Susi) FindStudent(facultyNumber int) (*Student, error)
    func (s *Susi) AddCourse(request []byte) error
    func (s *Susi) FindCourse(courseIdentifier string) (*Course, error)
    func (s *Susi) Enroll(request []byte)  error
    func (s *Susi) FindEnrollment(facultyNumber int, courseIdentifier string) (*Enrollment, error)

_Подробно описание на методите ще откриете по-долу._

### SusiError
Типа `SusiError` трябва да има елементи `Student` и `Course`, разбира се ако курса не е наличен, полето `Course` има nil стойност.

Ако грешката не е свързана със студент (например при `FindCourse`/`AddCourse`) е нормално `Student` да е празно.

Ако грешката не е свързана с курс (например при `FindStudent`/`AddStudent`) е нормално `Course` да е празно.

Този тип трябва да имплементира [`error`](http://golang.org/pkg/builtin/#error) интерфейса в Go.

### Course
Трябва да имплементира [`Stringer`](http://golang.org/pkg/fmt/#Stringer) и при извикване на метода String() да връща string подобен на:

    "AR Advanced Robotics"

тоест - _courseIdentifier courseName_

Този тип има следните полета:

* CourseName
* CourseIdentifier
* MinimumAcademicYear
* MastersOnly
* AvailablePlaces

### Student
Трябва да имплементира [`Stringer`](http://golang.org/pkg/fmt/#Stringer) и при извикване на метода String() връща string подобен на:

    "12345 Leeloo Dallas"

тоест - _ФН Име Фамилия_

Този тип има следните полета:

* FacultyNumber
* FirstName
* LastName
* AcademicYear
* Master

### Enrollment

* Course - Указател към обект за записания курс.
* Student - Указател към обект за записалия се студент.

## Други
За всеки друг тип, който дефиниране следва, че той не трябва да е публичен.

## Методи

### `func (s *Susi) AddCourse(request []byte) error`
Пример за request:

    {
        "course_name": "Advanced Robotics",
        "course_identifier": "AR",
        "minimum_academic_year": 3,
        "masters_only": false,
        "available_places": 2
    }

Добавя курса в `Susi`. Ако курса вече съществува връща грешка (вижте валидация).

_Използвайте [Unmarshal](http://golang.org/pkg/encoding/json/#Unmarshal)._

### `func (s *Susi) FindCourse(courseIdentifier string) (*Course, error)`
Връща указател към обект от тип `Course`, ако в системата има курс с подадения identifier.
Ако няма - връща грешка (виж валидация).

### `func (s *Susi) AddStudent(request []byte) error`
Пример за request:

    {
        "faculty_number": 12345,
        "first_name": "Leeloo",
        "last_name": "Dallas",
        "master": false,
        "academic_year": 2
    }

Добавя студента в `Susi`. Ако студента вече съществува връща грешка (вижте валидация).

_Използвайте [Unmarshal](http://golang.org/pkg/encoding/json/#Unmarshal)._

### `func (s *Susi) FindStudent(facultyNumber int) (*Student, error)`
Връща указател към обект от тип `Student`, ако в системата има студент с подадения факултетен номер.
Ако няма - връща грешка (виж валидация).

### `func (s *Susi) Enroll(request []byte)  error`
Пример за request:

    {
        "faculty_number": 12345,
        "course_identifier": "AR"
    }

Ако данните са валидни (вижте валидация) - записва студента за курс.

__Валидациите се правят в следния ред:__

* Има ли студент?
* Има ли курс?
* Дали вече няма enrollment?
* Има ли места в курса?
* Покриват ли се изискванията за курса?

_Използвайте [Unmarshal](http://golang.org/pkg/encoding/json/#Unmarshal)._

### `func (s *Susi) FindEnrollment(facultyNumber int, courseIdentifier string) (*Enrollment, error)`
При подадени `facultyNumber` и `courseIdentifier` трябва да върнете указател към обект от тип Enrollment, ако студента е записан за курса.
Преди да проверите дали има enrollment за този студент в този курс, проверете:

* Има ли студент?
* Има ли курс?
* Има ли enrollment?

## Валидация

* __Има вероятност да подадем невалиден json на `AddStudent`, `AddCourse` или `Enroll`. Ако получите грешка при Unmarshal-ването е редно просто да я върнете.__

* Ако при добавяне на студент, такъв вече съществува (със същия `facultyNumber`):
        "Студент с факултетен номер 12345 вече съществува!"

* Ако търсеният студент не е наличен:
        "Няма студент с факултетен номер 12345!"

* Ако при добавяне на курс, такъв вече съществува:
        "Курс с identifier AR вече съществува!"

* Ако търсеният курс не е наличен:
        "Няма курс с identifier - AR!"

* За всеки един курс има останали определен брой места, ако студент се опита да се запише за курс, за който вече няма свободни места - трябва да върнете грешка. Например:
        "Няма свободни места за курс с identifier - AR!"

* Ако студента е под нужния минимум (academic_year) или не покрива условието да е магистър, ще очакваме подобна грешка:
        "Leeloo Dallas не покрива изискванията за Advanced Robotics!"
    _По подразбиране магистрите удовлетворяват условието за академична година._

* Ако се опитаме за запишем студент, но вече имаме съществуващ Enrollment за този студент в съответния курс, се връща грешка, подобна на:
        "Студент с факултетен номер 12345 вече е записан за курс с identifier AR!"

* При FindEnrollment и липсващ Enrollment се връща грешка, подобна на:
        "Студент с факултетен номер 12345 не е записан за курса с identifier AR!"


При наличието на грешка трябва да прекратите изпълнението на метода и да я върнете като резултат. Грешките, които вие създавате да са от тип `SusiError`, който да имплементира `error` интерфейса в Go.

### Внимание!

Трябва да използвате съобщенията за грешка от примерите в предишната секция - "Валидация". Като в тях трябва да смените съответните данни - факултетен номер, identifier на курса, имена на студент или име на курс. Примери:

    student, err := mySusi.GetStudent(54321)
    if err != nil {
        fmt.Println(err) // Принтира "Няма студент с факултетен номер 54321!"
    }

    err = mySusi.AddCourse([]byte(`...,"course_identifier": "AI",...`))
    if err != nil {
        fmt.Println(err) // Принтира "Курс с identifier AI вече съществува!"
    }

В тестовете си ще проверяваме за точно тези стрингове и ако върнатите от вас се различават дори с един байт теста няма да бъде успешен.
