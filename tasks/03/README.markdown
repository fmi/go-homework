ExpireMap
===================

В тази задача ще си говорим за [кеширане](https://en.wikipedia.org/wiki/Cache_%28computing%29). Понякога е много полезно да се запази резултат от скъпа операция и вместо след това отново да се извърши скъпата операция да се използва вече запазеният резултат. Но света се променя и след време този резултат може да е невалиден. За това е хубаво да го пазим само определено време. Това време е различно за различните видове резултати. HTML-a на web страница може да се пази секунди или дори цели минути, но повече от това е прекалено много. Но IP-то на домейн може да се пази часове и дори дни, защото рядко се променя. Точно DNS-ите са хубав [пример как кеширането работи](https://en.wikipedia.org/wiki/Time_to_live#DNS_records) за да "ускори" многократно интернета ни.

Вашата задача ще е да направите тип `ExpireMap`, който да може да се използва за такова кеширане. Той трябва да е [key-value хранилище](https://en.wikipedia.org/wiki/Associative_array), подобен на `map` в Go, но всеки ключ ще има максимално време на живот. След като времето на живот на ключа изтече той и стойността му трябва да изчезнат от хранилището. Ключовете в хранилището са `string`, а стойностите от произволен тип (иначе казано - `interface{}`).

Хранилището трябва да може да бъде използваемо от много go рутини едновременно.

type ExpireMap
-------------------

Функциите и методите, които трябва да има вашето решение са следните

##### `func NewExpireMap() *ExpireMap`

Връща указател към нов обект от тип `ExpireMap`, готов за използване.

##### `func (em *ExpireMap) Set(key string, value interface{}, expire time.Duration)`

Добавя в хранилището нов ключ `key` със стойност `value` за `expire` време. 
[Duration](http://golang.org/pkg/time/#Duration) е дефиниран в пакета 
[time](http://golang.org/pkg/time/). `expire` време след добавянето на ключа той вече
трябва да спре да бъде в хранилището. Стойността може да е от всякакъв тип, докато ключът
е от тип `string`.

##### `func (em *ExpireMap) Get(key string) (interface{}, bool)`

Връща стойността на ключ и `true`, когато ключът е намерен. Когато ключът липсва или е изтекъл връща `nil` и `false`.

##### `func (em *ExpireMap) GetInt(key string) (int, bool)`

Връща стойността, отговаряща на този ключ и `true`, когато ключът е намерен и стойността му е от тип `int`. Когато ключът липсва, изтекъл е или стойността му не е от този тип методът трябва да върне нула и `false`.

##### `func (em *ExpireMap) GetFloat64(key string) (float64, bool)`

Връща стойността на този ключ и `true`, когато ключът е намерен и стойността му е от тип `float64`. Когато ключът липсва, изтекъл е или стойността му не е от този тип методът трябва да върне нула и `false`.

##### `func (em *ExpireMap) GetString(key string) (string, bool)`

Връща стойността на ключ и `true`, когато ключът е намерен и стойността му е от тип `string`. Когато ключът липсва, изтекъл е или стойността не е от тип `string` методът трябва да върне празен стринг и `false`.

##### `func (em *ExpireMap) GetBool(key string) (bool, bool)`

Връща стойността на ключ и `true`, когато ключът е намерен и стойността му е от тип `bool`. Когато ключът липсва, изтекъл е или стойността не е от тип `bool` методът трябва да върне `false` и `false`.

##### `func (em *ExpireMap) Expires(key string) (time.Time, bool)`

За определен ключ тази функция трябва да върне [времето](http://golang.org/pkg/time/#Time), когато той ще изтече. Втората върната стойност е `true` когато ключът е намерен в хранилището. Когато не е намерен функцията трябва да върне нулевото време и `false`. 

##### `func (em *ExpireMap) Delete(key string)`

Премахва ключа от хранилището. Когато няма такъв ключ методът не трябва да гърми.

##### `func (em *ExpireMap) Contains(key string) bool`

Казва дали ключ е в хранилището или не. Отново, ключове са в хранилището само ако техния
`expire` не е вече изтекъл.

##### `func (em *ExpireMap) Size() int`

Връща големината на хранилището. Големина е бройката неизтекли ключове, които са в него.

##### `func (em *ExpireMap) Increment(key string) error`

За ключ увеличава стойността му числово с 1 когато тази стойност е от тип `int` или стринг, представящ цяло число. Когато стойността е от тип, различен от тези неща функцията трябва да върне грешка, различна от `nil`. `float32` и `float64` не са целочислен типове. Времето на изтичане на ключа не трябва да се променя.

##### `func (em *ExpireMap) Decrement(key string) error`

Същото като `Increment`, но намалява стойността с 1 вместо да я увеличава.

##### `func (em *ExpireMap) ToUpper(key string) error`

Ако стойността за ключа `key` е от тип `string`, то всички букви в този `string` трябва да станат главни. При успех методът трябва да върне `nil`, а при невъзможност да изпълни задачата - грешка.
Пример за употреба:

```go
em := NewExpireMap()
em.Set("example", "creeper on the roof", 10 * time.Minute)
if err := em.ToUpper("example"); err != nil {
    // error converting the string. maybe it has expired?
}
upped, _ := em.GetString("example") // upped is "CREEPER ON THE ROOF"
```

##### `func (em *ExpireMap) ToLower(key string) error`

Същото като `ToUpper`, но всички букви трябва станат малки.

##### `func (em *ExpireMap) ExpiredChan() <- chan string`

Тази функция трябва да върне канал, по който да идват ключовете, които са изтекли. Те трябва да се пускат по канала в момента на изтичането си. Ключове, които са изтрити с `Delete` или `Cleanup` не трябва да се връщат по този канал. Единствено тези, които са достигнали до времето си на изтичане. Не е задължително през цялото време някой да чете от върнатия канал. Ключове добавени в инстанцията *след* извикването на `ExpireChan` също трябва да се пратят по канала, когато тяхното време дойде.

```go
em := NewExpireMap()
defer em.Destroy()

added := time.Now()

em.Set("key1", "val1", 50 * time.Millisecond)
em.Set("key2", "val2", 100 * time.Millisecond)

expires := em.ExpiredChan()

for i := 0; i < 2; i++ {
    key := <- expires
    fmt.Printf("%s expired after %s\n", key, time.Since(added))
}
```

##### `func (em *ExpireMap) Cleanup()`

Изчиства хранилището. След извикване на функцията в `ExpireMap`-a не трябва да се съдържа нито един ключ.

##### `func (em *ExpireMap) Destroy()`

Трябва да освободи всички ресурси, които е използвал този `ExpireMap`. Това означава изчистване на всички структури, затваряне на всички канали и спиране на всички горутини, свързани с нормалната му работа. След извикване на `Destroy` тази инстанция на `ExpireMap` няма да бъде повече използвана. Този метод би трябвало да се ползва по подобен начин:

```go
em := NewExpireMap()
defer em.Destroy()
```


## Пример

```go
cache := NewExpireMap()
defer cache.Destroy()

cache.Set("foo", "bar", 15*time.Second)
cache.Set("spam", "4", 25*time.Minute)
cache.Set("eggs", 9000.01, 3*time.Minute)

err := cache.Increment("spam")

if err != nil {
    fmt.Println("Sadly, incrementing the spam did not succeed")
}

if val, ok := cache.Get("spam"); ok {
    fmt.Printf("The value of spam should be the string 5: %s", val)
} else {
    fmt.Println("No spam. Have some eggs instead?")
}

if eggs, ok := cache.GetFloat64("eggs"); !ok || eggs <= 9000 {
    fmt.Println("We did not have as many eggs as expected. Have you considered our spam offers?")
}
```

Тестовете ни ще позволяват на вашите имплементации да се забавят до 10ms с изчистването на изтекли ключове. Много ще харесаме решения, които не хабят памет за да пазят вече изтекли ключове и стойности.
