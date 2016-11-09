# Concurrent Retry Executor

Напишете функция, която конкурентно и последователно изпълнява списък от подадени ѝ задачи и връща техните резултати като канал:

```
func ConcurrentRetryExecutor(tasks []func() string, concurrentLimit int, retryLimit int) <-chan struct {index int, result string}
```

Всяка задача от `tasks` е под формата на проста функция, която връща `string` като резултат. Ако върнатият `string` е празен, това се приема за грешка и трябва веднага да изпълните отново задачата. Всяка задача може да бъде изпълнена максимум `retryLimit` на брой пъти. В даден момент трябва да се изпълняват по `concurrentLimit` на брой задачи едновременно, ако това е възможно.

Резултатите от изпълнението на задачите (вкл. грешните празни отговори) заедно с поредните номера на самите задачи (техният индекс в `tasks`) трябва да бъдат изпращани по return канала. Този канал трябва да бъде създаден от `ConcurrentRetryExecutor` и да бъде затворен след като последната задача е приключила и нейният резултат е бил записан в канала.

Също така, трябва да е възможно няколко инстанции на `ConcurrentRetryExecutor` да работят едновременно, без да си пречат една на друга.

Ето как изглежда това на практика. Следният код:

```
first := func() string {
    time.Sleep(2 * time.Second)
    return "first"
}
second := func() string {
    time.Sleep(1 * time.Second)
    return "second"
}
third := func() string {
    time.Sleep(600 * time.Millisecond)
    return "" // always a failure :(
}
fourth := func() string {
    time.Sleep(700 * time.Millisecond)
    return "am I last?"
}

fmt.Println("Starting concurrent executor!")
tasks := []func() string{first, second, third, fourth}
results := ConcurrentRetryExecutor(tasks, 2, 3)
for result := range results {
    if result.result == "" {
        fmt.Printf("Task %d returned an error!\n", result.index+1)
    } else {
        fmt.Printf("Task %d successfully returned '%s'\n", result.index+1, result.result)
    }
}
fmt.Println("All done!")
```
би трябвало да изведе:

```
Starting concurrent executor!
Task 2 successfully returned 'second'
Task 3 returned an error!
Task 1 successfully returned 'first'
Task 3 returned an error!
Task 4 successfully returned 'am I last?'
Task 3 returned an error!
All done!
```
