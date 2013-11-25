MarkdownParser
===================

В тази задача ще се наложи да парсваме [markdown](http://daringfireball.net/projects/markdown/) документи. За целта ще ви
предоставим статия, написана на markdown, от която вие трябва да извлечете
заглавия, линкове, имена и друга основни за статията неща.


type MarkdownParser
-------------------

### `func NewMarkdownParser(text string) *MarkdownParser`

Връща указател към `MarkdownParser`.

### `func (mp *MarkdownParser) Headers() []string`

Връща всички H1 хедъри като слайс от стрингове. Обърнете внимание, че има
два начина, по които може да се дефинира H1 хедър.

### `func (mp *MarkdownParser) SubHeadersOf(header string) []string`

Приема H1 хедър и връща слайс от стринговете на всички H2 хедъри,
които са дефинирани след подадения H1 хедър и преди следващия такъв.
H3, H4, H5 и H6 хедъри не ни интересуват. Обърнете внимание, че има
два начина, по които може да се дефинира H2 хедър.

### `func (mp *MarkdownParser) Names() []string`

Връща всички имена в текста. За име приемаме две или повече поредни думи
с главни букви (допустими са тирета между две думи с или без интервал между тях),
игнорирайки първата дума в изречението. В този смисъл следните са имена:

* Иван Павлов
* Георги Кранев
* Иван Ковачев Павлов
* Едсон Арантес Ду Насименто - Пеле
* Mozilla Firefox

__Забележка__: От првилото да игнорирате първата дума в изречение следва,
че в изречението "Георги Кранев е много забавен." няма име.

### `func (mp *MarkdownParser) PhoneNumbers() []string`

Връща всички телефонни номера в текста.
Не се интересуваме от броя цифри, нито от префиските им
(демек не очакваме само български номера).
Преди цифрите може да има плюс и отваряща скоба, а между тях - интервали, скоби и тирета.
В този смисъл следните са телефонни номера:

* 0889123456
* +359889123456
* (089) 123-456
* 0 (889) 123 - 456
* +4531223 2332 123
* 123 3456 621

### `func (mp *MarkdownParser) Links() []string`

Връща всички изходящи линкове. Пишейки "линк" си мислим за
[Uniform resource locator](http://en.wikipedia.org/wiki/Uniform_resource_locator).

Т.е. очакваният формат е `scheme://domain:port/path?query_string#fragment_id`.
От тези само схемата, домейнът и пътят са задължителни(/ е валиден път).
Запознайте се с това къде какви символи са позволени.

### `func (mp *MarkdownParser) Emails() []string`

Връща всички Email адреси в текста. Не очакваме да напишете регулярен израз,
който да валидира абсолютно всички email адреси.

Валиден email адрес ще наричаме всяка последователност от символи, която:

* започва с малка, главна буква или цифра
* опционално могат да следват до 200 други букви, цифри, долни черти, плюсове, точки или тирета
* символ @
* валиден [домейн](http://en.wikipedia.org/wiki/Uniform_resource_locator#Syntax)

### `func (mp *MarkdownParser) GenerateTableOfContents() string`

Връща съдържанието на подадения текст, като номериран списък.
Всяка точка в номерирания списък е H1 хедър. Под-точка на тази точка е H2 хедър,
под-под-точка е H3 хедър на предния H2 хедър и т.н. до H6.

## Пример

За целите на примера, ще подадем условието на това домашно като вход:

    >>> mdParser := NewMarkdownParser(data)
    >>> mdParser.Headers()
    <<< []string{"type MarkdownParser", "Пример"}
    >>> mdParser.SubHeadersOf("Пример")
    <<< []string{}
    >>> mdParser.GenerateTableOfContents()
    <<< 1. MarkdownParser
    <<< 1.1 type MarkdownParser
    <<< 1.1.1 `func NewMarkdownParser(text string) *MarkdownParser`
    <<< 1.1.2 `func (mp *MarkdownParser) Headers() []string`
    <<< 1.1.3 `func (mp *MarkdownParser) SubHeadersOf(header string) []string`
    <<< 1.1.4 `func (mp *MarkdownParser) Names() []string`
    <<< 1.1.5 `func (mp *MarkdownParser) PhoneNumbers() []string`
    <<< 1.1.6 `func (mp *MarkdownParser) Links() []string`
    <<< 1.1.7 `func (mp *MarkdownParser) Emails() []string`
    <<< 1.1.8 `func (mp *MarkdownParser) GenerateTableOfContents() string`
    <<< 2. Пример
