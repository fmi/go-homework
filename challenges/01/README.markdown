# PubSub

Направете тип `PubSub` за изпращане на `string` съобщения до множество получатели. Този тип трябва да има два метода: `Subscribe()` и `Publish()`.

`Subscribe()` връща небуфериран канал за четене `<-chan string` и се използва за регистриране на нов "слушател". `Publish()` връща небуфериран канал за писане `chan<- string`, чрез който могат да се изпращат съобщения до всички регистрирани слушатели.

Напишете и функция `NewPubSub()`, която да връща инициализиран `*PubSub`, т.е. служи като конструктор на типа.

Пример:

```
ps := NewPubSub()
a := ps.Subscribe()
b := ps.Subscribe()
c := ps.Subscribe()
go func() {
    ps.Publish() <- "wat"
    ps.Publish() <- ("wat" + <-c)
}()
fmt.Printf("A recieved %s, B recieved %s and we ignore C!\n", <-a, <-b)
fmt.Printf("A recieved %s, B recieved %s and C received %s\n", <-a, <-b, <-c)
```

би трябвало да изведе:

```
A recieved wat, B recieved wat and we ignore C!
A recieved watwat, B recieved watwat and C received watwat
```

Уточнение: за улеснение, очакваме изпращането на едно съобщение или регистрирането на нов слушател да се случат чак след като всички регистрирани слушатели са прочели предишното съобщение. С други думи, не търсим ефективност, а коректност.
