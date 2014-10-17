Често срещани думи
=================

Да се направи функция `getMostCommonWords`, която приема 2 аргумента `text string` и `wordCountLimit int` и връща `[]string`. Функцията трябва да обходи подадения ѝ `text` и да върне списък (инициализиран слайс) с всички думи, преобразувани в lower case, които се срещат поне `wordCountLimit` пъти. Върнатият слайс трябва да е сортиран по азбучен ред.

В подадения текст, думите са разделени с интервали и могат да съдържат всякакви символи.

Example:

    text := "A Be Ce DE De a! oh? a oy! oY! OY!"
    mostCommonWords := getMostCommonWords(text, 2)

mostCommonWords трябва да е ["a", "de", "oy!"]
