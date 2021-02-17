package main

import (
	// "fmt"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	"encoding/xml"
	"encoding/json"
	"strings"
	"sort"
	"strconv"
	"io"
	"reflect"
	"testing"
	"time"
	// "os"
	// "url"
)
// код писать тут



type Dataset struct {
	XML xml.Name `xml:"root"`
	Rows    []Row   `xml:"row"`
}

type Row struct{//структура для записи значений  из xml
	Id int `xml:"id"`
	Age int `xml:"age"`
	First_name string `xml:"first_name"`
	Last_name string `xml:"last_name"`
	Gender string `xml:"gender"`
	About string `xml:"about"`
}


const filePath string = "dataset.xml"

func SearchServer(w http.ResponseWriter, r *http.Request){
	//План:
	
	//1. Достает из запроса параметры, проверяет токен
	//2. Читает xml и сохраняет каждую запись в структуру, создает из них слайс
	//3. Находит структуры соответсвующие параметрам запроса(подстрока в Name или About, с помощью strings.Contains),
	// и сохраняет их в слайс
	//4. Сортирует структуры по параметру OrderField
	//5. Преобразует данные в JSON и отправляет их обратно клиенту со статускодом
	
	//1.
    // Limit := r FormValue("limit")
	Query := r.FormValue("query")
	OrderField := r.FormValue("order_field")
	OrderBy,_ := strconv.Atoi(r.FormValue("order_by"))
	AccessToken := r.Header.Get("AccessToken")

	//2.	
	fileContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}

	xmlDoc := Dataset{} //все данные из xml тут
	err = xml.Unmarshal(fileContents, &xmlDoc)//за
		if err!= nil {
				panic(err)
			}
	
	//3.
	//достаем из струтуры xmlDoc поле Rows,в нем достаем все значения и перезаписываем 
	//в структуру с полем Name
	//Далее каждую измененную структуру записываем в newRows
		type newRow struct { // структура для перезаписи исходной структуры Row
			Id int
			Name string
			Age int
			About string
			Gender string
		}
		
		newRows := []newRow{}//слайс структур с полем Name

	for _,row := range xmlDoc.Rows{
		user := newRow{ //перезаписываем структуру
			Id: row.Id,
			Name: row.First_name+" "+row.Last_name,
			Age: row.Age,
			About: row.About,
			Gender: row.Gender,
		}

		newRows = append(newRows, user) 
	}

	resultQuery := []newRow{}//сохраняет результат запроса
	for _,user := range newRows {//достаем каждого пользователя по отдельности и ищем подстроку
		if (strings.Contains(user.Name, Query)||strings.Contains(user.About, Query)){
			resultQuery = append(resultQuery, user)//добавляем юзера в результат, если найдена подстрока
		}
		continue
	}
    //4. 
	if OrderBy == -1{
		switch OrderField { //сортировка в зависимости от значения OrderField и OrderBy -1
		case "Id":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Id > resultQuery[j].Id})
		case "Age":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Age > resultQuery[j].Age})
		case "Name":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Name > resultQuery[j].Name})
		case "":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Name > resultQuery[j].Name})
		}
	}	

	if OrderBy == 1{
		switch OrderField { //сортировка в зависимости от значения OrderField и OrderBy 1
		case "Id":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Id < resultQuery[j].Id})
		case "Age":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Age < resultQuery[j].Age})
		case "Name":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Name < resultQuery[j].Name})
		case "":
			sort.Slice(resultQuery, func(i,j int)bool {return resultQuery[i].Name < resultQuery[j].Name})
		}
	}	
	if OrderBy == 0{	//сортировка в зависимости от значения  OrderBy 0
		
	}	
		//5.
	if AccessToken != "token"{//TestStatusUnauthorized
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w,`{"status": 401}`)	
	} 
	
	if 	AccessToken == "token"{	
		switch Query{
		case "Dillard"://TestClient общий тест если все хорошо
			jsonDoc,_ := json.Marshal(resultQuery)
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, string(jsonDoc))

		case "Mccoy"://TestClient общий тест, но с кривым json
			//jsonDoc,_ := json.Marshal(resultQuery)
			w.WriteHeader(http.StatusOK)
			io.WriteString(w,`{User": "Error"`)	
		
		case "Boyd"://TestOffset TestLimit
			jsonDoc,_ := json.Marshal(resultQuery)
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, string(jsonDoc))
		
		case "ServerError": //TestStatusInternalServerError
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w,`{"status": 500}`)	

		case "BadRequest"://TestBadRequest
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w,`{"Error": "ErrorBadOrderField"}`)	
		
		case "BadRequest2"://TestBadRequest
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w,`{Error": "ErrorBadOrderField"`)//если внутри кривой json	
		
		case "BadRequest3"://TestBadRequest
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w,`{"Error": "unknownError"}`)	
		case "Sleep":
			time.Sleep(3 * time.Second )
			w.WriteHeader(http.StatusOK)
			io.WriteString(w,``)

		}
	
	}

}

//Тест
// 1. Определяем тип TestCase
// 2. Пишем тестовую функцию и передаем туда обработчик SearchServer
// 3. В тестовой функции создаем тестовый сервер 
// 4. Создаем объект с типом SearchClient для которого определяем токен и адрес тестового сервера
// 5. 


type TestCase struct {//структура для тестового кейса
	Request SearchRequest
	Result  SearchResponse
	IsError bool
}


func TestClient(t *testing.T){//общий тест. задействует if len(data) == req.Limit 
	
	casetest := TestCase{
			Request: SearchRequest{
				Limit: 1,
				Offset: 0,
				Query:"Dillard",
				OrderField:"",
				OrderBy:1,
			},
			Result: SearchResponse{
				Users: []User{
					User{
						Id: 17,
						Name: "Dillard Mccoy",
						Age: 36,
						About:"Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n",
						Gender: "male",
					},
					
				},
			},
			IsError: false,
	}	

	//создаем тестовый сервер, передаем в него SearchServer - функцию,
	// которая будет обрабатывать запрос от Client.go

	server := httptest.NewServer(http.HandlerFunc(SearchServer))
		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		result,_:= sc.FindUsers(casetest.Request)
			
			eq:= reflect.DeepEqual(casetest.Result.Users, result.Users)//сравнивает ожидаемый результат с фактическим
			if  !eq {
				t.Errorf("Ожидается %#v,\n получено %#v ", casetest.Result, result,)
				// fmt.Println("Тесткейс",numCase,"прошел","\n","Фактический:",result,len(result.Users),"\n","Ожидаемый:",casetest.Result, len(casetest.Result.Users))
			}		
			// fmt.Println("Ожидаемое и фактическое значение совпали")
			server.Close()	
}

func TestClientBadJson(t *testing.T){//общий тест. задействует ошибку распаковки JSON
	
	casetest := TestCase{
			Request: SearchRequest{
				Limit: 1,
				Offset: 0,
				Query:"Mccoy",
				OrderField:"",
				OrderBy:1,
			},
			Result: SearchResponse{
				Users: []User{
					User{
						Id: 17,
						Name: "Dillard Mccoy",
						Age: 36,
						About:"Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n",
						Gender: "male",
					},
					
				},
			},
			IsError: false,
	}	

	//создаем тестовый сервер, передаем в него SearchServer - функцию,
	// которая будет обрабатывать запрос от Client.go

	server := httptest.NewServer(http.HandlerFunc(SearchServer))
		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		_,err:= sc.FindUsers(casetest.Request)
			if  err == nil {
				t.Errorf("Ожидается cant unpack result json ,\n получено %#v ", casetest.Result)
				// fmt.Println("Тесткейс",numCase,"прошел","\n","Фактический:",result,len(result.Users),"\n","Ожидаемый:",casetest.Result, len(casetest.Result.Users))
			}		
			// fmt.Println("Ожидаемое и фактическое значение совпали")
			server.Close()	
}

func TestLimit(t *testing.T){

		casetest := TestCase{
        	Request:SearchRequest{
					Limit: -2,
					Offset: 0,
					Query:"Boyd",
					OrderField:"",
					OrderBy:1,
				},
			Result: SearchResponse{
				Users: nil,
			},
			IsError: true,
		}
		

		server := httptest.NewServer(http.HandlerFunc(SearchServer))
			sc := &SearchClient{
			AccessToken: "token",
			URL: server.URL,
			}
			_,err:= sc.FindUsers(casetest.Request)
				if err == nil {
					t.Errorf(" Ожидается limit must be > 0,\n получено %#v", nil)
				}
				// fmt.Println(err)
				server.Close()	
}

func TestLimit25(t *testing.T){
	
	casetest := TestCase{
		Request:SearchRequest{
				Limit: 27,
				Offset: 0,
				Query:"Boyd",
				OrderField:"",
				OrderBy:1,
			},
		Result: SearchResponse{
				Users: []User{
					User{
					Id: 0,
					Name: "Boyd Wolf",
					Age: 22,
					About:"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
					Gender: "male",
					},
				},
			},	
			IsError: false,
	}	
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		result,_:= sc.FindUsers(casetest.Request)

		eq:= reflect.DeepEqual(casetest.Result.Users, result.Users)
				if !eq {
					t.Errorf("Ожидается Boyd Wolf")
				}
    }				

func TestOffset(t *testing.T){

	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 3,
					Offset: -1,
					Query:"Boyd",
					OrderField:"",
					OrderBy:1,
				},
			// Result: SearchResponse{
			// 		Users: []User{
			// 		},
			//	},
			IsError: true,

		}
		
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		result,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено %#v ",  casetest.Result, result)
			}
			// fmt.Println(err)
			server.Close()	
}

func TestStatusUnauthorized(t *testing.T){
	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 2,
					Offset: 0,
					Query:"Boyd",
					OrderField:"",
					OrderBy:1,
				},
				Result: SearchResponse{
					Users: nil,
				},
				IsError: true,

		}
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "BadToken",
		URL: server.URL,
		}

		_,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено  err == nil",  casetest.Result)
			}
			// fmt.Println(err)
	
			server.Close()	
}

func TestStatusInternalServerError(t *testing.T){
	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 2,
					Offset: 0,
					Query:"ServerError",
					OrderField:"",
					OrderBy:1,
				},
				Result: SearchResponse{
					Users: nil,
				},
				IsError: true,

		}
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		_,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено err == nil ",  casetest.Result)
			}
			// fmt.Println(err)
			server.Close()	
}

func TestStatusBadRequest(t *testing.T){
	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 2,
					Offset: 0,
					Query:"BadRequest",
					OrderField:"",
					OrderBy:1,
				},
				Result: SearchResponse{
					Users: nil,
				},
				IsError: true,

		}
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		_,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено err == nil ",  casetest.Result)
			}
			// fmt.Println(err)
			server.Close()	
}
func TestStatusBadRequest2(t *testing.T){
	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 2,
					Offset: 0,
					Query:"BadRequest2",
					OrderField:"",
					OrderBy:1,
				},
				Result: SearchResponse{
					Users: nil,
				},
				IsError: true,

		}
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		_,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено err == nil ",  casetest.Result)
			}
			// fmt.Println(err)
			server.Close()	
}

func TestStatusBadRequest3(t *testing.T){
	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 2,
					Offset: 0,
					Query:"BadRequest3",
					OrderField:"",
					OrderBy:1,
				},
				Result: SearchResponse{
					Users: nil,
				},
				IsError: true,

		}
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: server.URL,
		}

		_,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено err == nil ",  casetest.Result)
			}
			// fmt.Println(err)
			server.Close()	
}


func TestClientDo(t *testing.T){//тест client.Do(searcherReq)

	casetest:=TestCase{
        Request:SearchRequest{
					Limit: 2,
					Offset: 0,
					Query:"Wolf",
					OrderField:"",
					OrderBy:1,
				},
				Result: SearchResponse{
					Users: nil,
				},
				IsError: true,

		}
		server := httptest.NewServer(http.HandlerFunc(SearchServer))

		sc := &SearchClient{
		AccessToken: "token",
		URL: "",//не задаем URL. это провоцирует ошибку в client.Do 
		}

		_,err:= sc.FindUsers(casetest.Request)

			if err == nil && casetest.IsError{
				t.Errorf(" Ожидается %#v,\n получено err == nil ",  casetest.Result)
			}
			// fmt.Println(err)
			server.Close()	

	}

	
	func SleepDummy(w http.ResponseWriter, r *http.Request){
		time.Sleep(time.Second * 6)
		return
	}

	func TestClientDoTimeout(t *testing.T){//тест client.Do(searcherReq)
	
		casetest:=TestCase{
			Request:SearchRequest{
						Limit: 2,
						Offset: 0,
						Query:"Sleep",
						OrderField:"",
						OrderBy:1,
					},
					Result: SearchResponse{
						Users: nil,
					},
					IsError: true,
			}
			server := httptest.NewServer(http.HandlerFunc(SleepDummy))
			
			//client  = &http.Client{Timeout:time.Nanosecond }
		

			sc := &SearchClient{
			AccessToken: "token",
			URL: server.URL,//не задаем URL. это провоцирует ошибку в client.Do 
			}
	
			_,err:= sc.FindUsers(casetest.Request)
	
				if err == nil && casetest.IsError{
					t.Errorf(" Ожидается %#v,\n получено err == nil ",  casetest.Result)
				}
				// fmt.Println(err)
				server.Close()	
	
		}

