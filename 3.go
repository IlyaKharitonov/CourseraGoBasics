package main

import (
	"io"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"bytes"
	"sort"
	"sync"
)

func FastSearch(out io.Writer) {

		//заменил функции os.Open ioutil.ReadAll одной 
		fileContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
		//режет общий слайс байт на более мелкие слайсы там, где "\n"
		lines := bytes.SplitAfter(fileContents, []byte("\n"))
		
		//заменил интерфейсы структурой. поля соответствуют полям из документа
		//Index для последующей сортировки в слайсе
		//отпала необходимость в постоянных преобразованиях типа
		type data struct{
			Browsers []string
			Company string
			Country string
			Email string
			Job string
			Name string
			Phone string
			Index int
		}

		
		var mu sync.Mutex	
		var wg1 sync.WaitGroup

		//определил емкость слайса. соответствует значению из текстового документа
		const iterNum = 1000
		users := make([]data, 0, iterNum)	

		for i, line := range lines {
			user := data{}
			wg1.Add(1)

			//распараллелил вычисление json
			go func(line []byte, index int){ 
				err := json.Unmarshal(line, &user)
					if err != nil {
						panic(err)
					}
				user.Index = index	
					
				mu.Lock()
				users = append(users, user)
				mu.Unlock()
				wg1.Done()	
			}(line,i)		
		}
		wg1.Wait()

		//сортирую распакованные json'ы по полю Index в структуре
		sort.Slice(users, func(i, j int) bool { return users[i].Index < users[j].Index })
			
			seenBrowsers := make([]string,0,114)
			uniqueBrowsers := 0
			foundUsers := ""
			
		for _, user := range users {
		
				isAndroid := false
				isMSIE := false
						
				for _,browser := range user.Browsers {
						
						if ok := strings.Contains(browser, "Android"); ok  {
							isAndroid = true
							notSeenBefore := true
							for _, item := range seenBrowsers {
								if item == browser {
									notSeenBefore = false
								}
							}
							if notSeenBefore {
								seenBrowsers = append(seenBrowsers, browser)
								uniqueBrowsers++
							}
						}
					}
				
					for _,browser := range user.Browsers {
						
						if ok := strings.Contains(browser, "MSIE"); ok{
							isMSIE = true
							notSeenBefore := true
							for _, item := range seenBrowsers {
								if item == browser {
									notSeenBefore = false
								}
							}
							if notSeenBefore {
								seenBrowsers = append(seenBrowsers, browser)
								uniqueBrowsers++
							}
						}
					}
					
					
				if !(isAndroid && isMSIE) {
					continue
				}
				
				// log.Println("Android and MSIE user:", user["name"], user["email"])
				email := strings.Replace(user.Email, "@", " [at] ", -1)
				foundUsers += fmt.Sprintf("[%d] %s <%s>\n", user.Index, user.Name, email)
				
		}
		
		fmt.Fprintln(out, "found users:\n"+foundUsers)
		fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
	
}
func main(){
	
}

	