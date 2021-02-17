package main
 
 import (
	 "fmt"
	 "io/ioutil"
	"strings"
	 )

func dir (dirName string, last bool){

//блоки графического оформления консоли
element1 := "├───"
element2 := "└───"
element3 := " │"
emptyElement := "  "
tab := "\t"
var varElement string

//Сканирует директорию и помещает название файлов и папок в срез
files,_ := ioutil.ReadDir(dirName)

//функция определяет длину среза files и индекс последнего элемента
//для которого нужен тупиковый символ "└───"
endElem := len(files)-1
//достаем из среза files по одному все элементы
for indexFile, file := range(files){
//формируем для них отступ
//узнаем ширирну отступа 
//единица ширины = │ + \t + └───
widthIdent := len(strings.Split(dirName,"/"))

//задаем отступ
ident := ""

//цикл, который плюсует графические элементы
//число итераций равно widthIdent
	for i := 1; i<widthIdent; i++{
				
			if (i <= 1) || (last == false){
				//вставляется | элемент
				varElement = element3	
			}else{
				//вставляется пустой элемент
				varElement = emptyElement
			}
			
	ident = ident + varElement + tab	
	}	

	if file.IsDir(){
			if indexFile == endElem {
				//печатает название директории с └───
				fmt.Println(ident, element2, file.Name())
				//сообщает,что это последняя директория
				last = true
				
			}else{
				//печатает название директории с ├───
				fmt.Println(ident,element1,file.Name())
				//сообщает,что это не последняя директория
				last = false
			}
			//формирует адрес для рекурсивного запуска функции
			adress := dirName + "/"+file.Name()
			//рекурсивный запуск
			dir(adress,last)
		}else{
			if indexFile == endElem{
				//печатает название файла с └─── 
				fmt.Println(ident,element2,file.Name())	
				 
			}else{
				//печатает название файла с ├───
				fmt.Println(ident,element1,file.Name())
			}
		}
	}
}

func main () {
	dir("testdata",false)
} 
