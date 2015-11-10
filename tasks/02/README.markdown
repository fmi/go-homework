# Ordered Log Drainer

Напишете функция `OrderedLogDrainer`, която приема канал от `string` канали `logs` като аргумент и връща прост `string` канал `result`. Дефиницията на функцията изглежда така:

```go
func OrderedLogDrainer(logs chan (chan string)) chan string
```

Всеки канал, който бъде пратен по `logs`, представлява прост лог с текстови съобщения. Работата на `OrderedLogDrainer` е да прочете всички пристигащи `string` съобщения от всички канали в `logs`, да ги подреди и, обединявайки ги в един канал, да ги изпрати по `result`. Ето правилата:

 - `OrderedLogDrainer` трябва винаги да "консумира" входящите съобщения по всички `logs` канали, така че писането в който и да е от тях да не блокира.
 - Всяко получено от каналите в `logs` съобщение трябва да бъде изпратено по канала `result`. Съобщенията, изпратени в `result` трябва да са подредени в определен ред. Първо по номера на техния канал в `logs` и след това по времето на изпращането им. Тоест всички съобщения, пристигнали от първия канал в `logs`, се пращат в `result` преди съобщенията от втория канал в `logs`. За информация по тази точка, вижте примерите.
 - Всяко съобщение трябва да бъде трансформирано малко. Искаме в началото му да се добави като prefix номера на log-а, от който е дошло, последван от табулация (`\t`). Номерата на логовете започват от 1. Номерът на лога е поредноста му на пристигане в `logs` канала.
 - Каналът `logs` и всички получени от него канали се затварят от извикващите функцията. В случая, това ще бъдат нашите тестове. Когато `logs` и всеки изпратен по него канал от string-ове бъдат затворени и всички съобщения в `result` са прочетени, `OrderedLogDrainer` трябва да затвори `result`.
 - `OrderedLogDrainer` трябва максимално бързо да изпраща съобщения по `result`. Това означава, че съобщенията от първия log в `logs` винаги максимално бързо се изпращат по `result` - не трябва да чакате да се получат всички съобщения от този log (т.е. да се затвори), за да започнете да ги изпращате по `result`. След като първия log бъде затворен, същото правило важи и за втория и т.н.
 - За улеснение, по всеки канал от `logs` ще бъдат изпратени максимум 100 съобщения.

Ето как изглежда задачата на практика. Следното парче код:
```go
logs := make(chan (chan string))
orderedLog := OrderedLogDrainer(logs)

first := make(chan string)
logs <- first
second := make(chan string)
logs <- second

first <- "test message 1 in first"
second <- "test message 1 in second"
second <- "test message 2 in second"
first <- "test message 2 in first"
first <- "test message 3 in first"
// Print the first message now just because we can
fmt.Println(<-orderedLog)

third := make(chan string)
logs <- third

third <- "test message 1 in third"
first <- "test message 4 in first"
close(first)
second <- "test message 3 in second"
close(third)
close(logs)

second <- "test message 4 in second"
close(second)

// Print all the rest of the messages
for logEntry := range orderedLog {
    fmt.Println(logEntry)
}
```

трябва да работи и да изведе на конзолата:
```
1	test message 1 in first
1	test message 2 in first
1	test message 3 in first
1	test message 4 in first
2	test message 1 in second
2	test message 2 in second
2	test message 3 in second
2	test message 4 in second
3	test message 1 in third
```
