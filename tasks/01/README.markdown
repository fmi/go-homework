# Poor Man's Currying

В тази задача ще трябва да напишете няколко функции, които генерират други функции. Това е много подобно на [currying](https://en.wikipedia.org/wiki/Currying) с това, че фиксирате няколко от аргументите на операцията и след това използвате функции с по - малък брой аргументи.


## Repeater

Функция, която приема string `s` и разделител `sep` и връща функция, която построява повторенията определен брой пъти и ги връща.

```
func Repeater(s, sep string) func (int) string
```

Която може да се използва по следния начин

```
Repeater("foo", ":")(3) // foo:foo:foo
```

## Generator

Функция, създава "генератор" функция за `int` числа.

```
func Generator(gen func (int) int, initial int) func() int
```

На `Generator` се подава `gen` функция и първоначалната стойност в поредицата - `initial`. `gen` взима като аргумент предишно изчислената стойност и връща следващата. 

Примерна употреба

```
counter := Generator(
    func (v int) int { return v + 1 },
    0,
)
power := Generator(
    func (v int) int { return v * v },
    2,
)

counter() // 0
counter() // 1
power() // 2
power() // 4
counter() // 2
power() // 16
counter() // 3
power() // 256
```

## MapReducer

Функция, която създава map reducer функция за `int` аргументи с подадени [map](https://en.wikipedia.org/wiki/Map_(higher-order_function)) функция, [reduce](https://en.wikipedia.org/wiki/Fold_(higher-order_function)) функция и първоначална стойност `initial` за reduce функцията.

```
func MapReducer(mapper func (int) int, reducer func (int, int) int, initial int) func (...int) int
```

Която може да бъде използвана по следния начин

```
powerSum := MapReducer(
    func (v int) int { return v * v },
    func (a, v int) int { return a + v },
    0,
)

powerSum(1, 2, 3, 4) // 30
```

Аргументите на `reducer` трябва да се подават от ляво на дясно. Иначе казано - напишете [left-fold](https://en.wikipedia.org/wiki/Fold_(higher-order_function)#On_lists) във вашата имплементация.

## Напомняне

Не забравяйте, че предадените решения трябва да са в пакет `main`. [Форматирайте кода си с gofmt](https://blog.golang.org/go-fmt-your-code).
