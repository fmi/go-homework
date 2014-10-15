Functions
================

Имаме функции и понеже имаме функции може да
пробваме няколко полезни неща, които принципно
 се свързват със функционалното програмиране.

Вие ще имплементирате няколко често срещани
и много полезни функции които имат много имена,
но за целите на домашното ще ги наричаме map, filter, reduce, all и any.

Всяка една от тях приема няколко параметъра като първия от тях винаги ще е `[]string` - които ще бъдат вашите данни.

Map
===

    func Map(data []string, mapper func(string) string) []string

Map мапва данните чрез втория ѝ параметър (мапъра) от едни стойности в други.

Пример:
-------
    length := func(s string) string {
        return strconv.Itoa(len(s)) // дължината на подадения string
    }
    result := Map([]string{"1", "two", "three", "four"}, length)
    // []string{"1", "3", "5", "4"}

Filter
======

    func Filter(data []string, predicate func(string) bool) []string


Filter филтрира елементите като за всеки елемент извиква predicate върху него и връща само тези за които получи true.

Пример:
-------
    result := Filter([]string{"1", "two", "three", "four"}, func(s string) bool {
        return len(s) == 3
    })
    // []string{"two"}

Reduce
======

    func Reduce(data []string, combinator func(string, string) string) string

Reduce е по забавен защото връща само една стойност независимо колко много елемента има. Връщаната стойност се получава като последователно извикваме combinator функцията върху до сега получената стойност и следващия елемент. За начална стойност ползвайте първата стойност от масива.

Пример:
-------

    concat := func(s1 string, s2 string) string {
        return s1 + s2
    }
    result := Reduce([]string{"1", "two", "three", "four"}, concat)
    // "1twothreefour"


Any
===

    func Any(data []string, predicate func(string) bool) bool

Any връща true ако за кой да е елемент predicate връща true.


All
===
    func All(data []string, predicate func(string) bool) bool

All връща true ако за всеки елемент predicate връща true.

Забележки:
----------
Както надявам се вече сте предположили Тези функции могат да се комбинарат лесно:

    result := Reduce(Filter(Map([]string{"1", "two", "three", "four"}, length), func(s string) bool {
        return s != "3"
    }), concat)
    result // "154"
