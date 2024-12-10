package enum

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"sync"

	"github.com/lexcf/o365sprayer/o365sprayer/logging"
	"github.com/lexcf/o365sprayer/o365sprayer/constants"
	"github.com/fatih/color"
)

var countManaged = 0

func counterManaged() {
	countManaged += 1
}

func ValidateEmailManagedO365(command string, email string, file *os.File,) {

	getCredentialTypeRequestJSON := constants.GetCredentialTypeRequestJSON{
		Username: email,
	}
	jsonData, _ := json.Marshal(getCredentialTypeRequestJSON)
	client := &http.Client{}
	req, err := http.NewRequest("POST", constants.GET_CREDENTIAL_TYPE, bytes.NewBuffer(jsonData))
	req.Header.Add("User-Agent", constants.USER_AGENTS[rand.Intn(len(constants.USER_AGENTS))])
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var getCredentialTypeResponseJSON constants.GetCredentialTypeResponseJSON
	json.Unmarshal(body, &getCredentialTypeResponseJSON)
	if getCredentialTypeResponseJSON.IfExistsResult == 0 {
		go counterManaged()
		color.Green("[*] Valid User : " + email)
		logging.LogEnumeratedAccount(file, email)
	}
	if command == "standalone" && getCredentialTypeResponseJSON.IfExistsResult != 0 {
		color.Red("[-] Invalid User : " + email)
	}
}

func EnumEmailsManagedO365(domainName string, command string, email string, filepath string, delay float64, threads int) {
	semaphore := make(chan struct{}, threads)
	var wg sync.WaitGroup
	semaphore <- struct{}{}
	wg.Add(1)
	go func() {
		defer func() {
			<-semaphore
			wg.Done()
		}()

		color.Yellow("[+] Enumerating Valid O365 Emails - Managed")
		timeStamp := strconv.Itoa(time.Now().Year()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(int(time.Now().Hour())) + strconv.Itoa(int(time.Now().Minute())) + strconv.Itoa(int(time.Now().Second()))
		fileName := domainName + "_enum_" + timeStamp
		logFile, err := os.OpenFile((fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Could not open " + fileName)
			return
		}
		defer logFile.Close()
		if command == "standalone" {
			ValidateEmailManagedO365(command, email, logFile)
		}
		if command == "file" {
			file, err := os.Open(filepath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)

			// Создаем канал с размерами буфера, чтобы ограничить количество одновременно работающих горутин
			concurrentLimit := threads
			sem := make(chan struct{}, concurrentLimit)

			// Для ожидания завершения всех горутин
			var wg sync.WaitGroup

			// Считываем строки из файла
			for scanner.Scan() {
				// Запускаем горутину для каждой строки
				wg.Add(1)
				go func(email string) {
					defer wg.Done()

					// Пытаемся захватить место в канале (блокирует, если превышен лимит)
					sem <- struct{}{}
					defer func() { <-sem }() // Освобождаем место в канале после завершения работы горутины

					// Выполняем вашу валидацию
					ValidateEmailManagedO365(command, email, logFile)

				}(scanner.Text())

				// Небольшая пауза для имитации задержки, можно настроить в зависимости от необходимости
				//time.Sleep(10 * time.Millisecond)
				time.Sleep(time.Duration(delay))
			
			}
			
			wg.Wait()

			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
			if countManaged > 0 {
				color.Yellow("[+] " + strconv.Itoa(countManaged) + " Valid O365 Emails Found !")
			} else {
				color.Red("[-] No Valid O365 Email Found !")
			}
		}

	}()

	wg.Wait()
}
