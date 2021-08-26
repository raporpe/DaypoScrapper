
# DaypoScrapper

Daypo.com is a plaform where users can upload custom made test so that other people can learn.

<img alt="daypo homepage" src="https://user-images.githubusercontent.com/6137860/130904854-2991469e-2eb1-4f4d-975a-f76cc43a8c3a.png">
<img alt="daypo test interface" src="https://user-images.githubusercontent.com/6137860/130906571-534e586a-1ef1-4aba-95e7-d1284b98be6e.png">

This website was designed many years ago and has an important flaw: the search system only allows searching tests by title but not by description. This makes finding tests very difficult since the test titles usually do not contain useful information and many times is completly unrelated to the actual content.

This web scrapper written in Go gathers information about the tests (such as the description) and stores it into a database so that thay can be searched with a query.
