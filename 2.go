package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//принимает в себя функции с типом job
func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup
	//канал входа в SingleHash
	in := make(chan interface{})
	//запускает каждую функцию конвеера по отдельности
	for _, funcWorker := range jobs {
		//выход из горутины, он же вход для следующей ступени конвеера
		out := make(chan interface{})
		wg.Add(1)
		go func(worker job, in chan interface{}, out chan interface{}) {
			worker(in, out)
			close(out)
			wg.Done()
		}(funcWorker, in, out)
		//обновляет вход для следующей функции
		in = out
	}
	//ожидает завершения всех горутин цикла
	wg.Wait()
}

//считает хэш сумму от каждого элемента, затем
//отправляет каждый результат по выходящему каналу
func SingleHash(in chan interface{}, out chan interface{}) {
	var wg sync.WaitGroup
	// позволяет изолировать кусок кода, осуществляет доступ к нему по очереди
	var mutex sync.Mutex

	for d := range in {
		wg.Add(1)
		//каждое входящее значение обрабатывается своей горутиной
		go func(d interface{}) {
			//приведение типа
			intData, ok := d.(int)
			if !ok {
				panic("Hallo, is panic SH")
			}
			// конвертация числа в строку
			myStr := strconv.Itoa(intData)
			//каналы по которым горутины будут обмениваться между собой данными
			chanCrc32 := make(chan string)
			chanMd5 := make(chan string)
			chanCrc32Md5 := make(chan string)
			//каждая функция хэш суммы обрабатывается своей горутиной,
			//а синхронизация происходит с помощью каналов
			go func(myStr string) {
				chanCrc32 <- DataSignerCrc32(myStr)
			}(myStr)

			go func(myStr string) {
				//мутекс блокирует одновременное использование несколькими горутинами
				mutex.Lock()
				chanMd5 <- DataSignerMd5(myStr)
				mutex.Unlock()
			}(myStr)

			go func(in chan string) {
				chanCrc32Md5 <- DataSignerCrc32(<-in)
			}(chanMd5)

			//вычисление итоговой конкатенации
			hash := <-chanCrc32 + "~" + <-chanCrc32Md5
			fmt.Println("Посчитал SingleHash ", myStr, " - ", hash)
			// отправляет результат по каналу
			out <- hash
			wg.Done()
		}(d)
	}
	wg.Wait()
}

//для каждого значения из синглхэш запускается своя рутина
//в этих горутинах запускается еще по 6 горутин для каждого значения счетчика
//в конце значения сортируются
func MultiHash(in chan interface{}, out chan interface{}) {
	var wg1 sync.WaitGroup
	for hash := range in {
		wg1.Add(1)
		go func(hash interface{}) {
			//приведение типа
			myStr, ok := hash.(string)
			if !ok {
				panic("Hallo, is panic MH")
			}
			//канал для отправки резузальтатов в виде структур
			//с сохранением порядкового номера для дальнейшей сортировки
			chanRes := make(chan struct {
				value string
				index int
			}, 6)
			//группа ожидания для синх. материнской и дочерних горутин
			var wg sync.WaitGroup
			wg.Add(6)

			fmt.Println("Считаю MultiHash значения", myStr)

			for i := 0; i <= 5; i++ {
				go func(i int, in chan struct {value string; index int}) { 
				//сообщает Wait о завершении рабоы горутины
				defer wg.Done()
				//преобразование числа в строку
				th := strconv.Itoa(i)
				//конкатенация входящего числа и входящего хэш-значения
				hashth := th + myStr
				//хэш
				crc32 := DataSignerCrc32(hashth)
				//порядковый номер
				index := i
				//структура для отправки
				var hashResult struct {
					value string
					index int
				}
				hashResult.value = crc32
				hashResult.index = index
				//передаем в канал  все значения
				in <- hashResult }(i, chanRes)
			}

			wg.Wait()
			// слайс, который собирает все результаты из горутин выше
			var crc32Slice []struct {
				value string
				index int
			}
			for elem := range chanRes {
				// закрытие канала
				if len(chanRes) == 0 {
					close(chanRes)
				}
				// добавляет хэш-значения из канала в конец слайса пока канал не опустеет
				crc32Slice = append(crc32Slice, elem)
			}
			//упорядоченный слайс структур в порядке возрастания счетчика
			sort.Slice(crc32Slice, func(i, j int) bool { return crc32Slice[i].index < crc32Slice[j].index })

			result := ""
			//достаем из структур и отправляем все хэш-значения в порядке возрастания счетчика
			for i, elem := range crc32Slice {
				fmt.Println(myStr, "th =", i, " ", elem.value)
				result = result + elem.value
			}
			fmt.Println("Конк. всех элементов цикла от", myStr, ": ", result)
			out <- result
			wg1.Done()
		}(hash)
	}
	wg1.Wait()
}

// собирает вместе результаты вычислений MH, сортирует, добавляет подчеркиваия между результатами
func CombineResults(in chan interface{}, out chan interface{}) {

	var combSlice []string
	//собираем все результаты в одном срезе из входящего канала
	for elem := range in {
		//приведение типа
		myStr, ok := elem.(string)
		if !ok {
			panic("Hallo, is panic CR")
		}
		combSlice = append(combSlice, myStr)
	}
	//сортируем слайс
	sort.Strings(combSlice)
	//добавляем подчеркивания
	resSlice := strings.Join(combSlice, "_")
	fmt.Println(resSlice)
	out <- resSlice
}

func main() {

	// ExecutePipeline(SingleHash(in,out),MultiHash(in,out),CombainResults(in,out))
}
