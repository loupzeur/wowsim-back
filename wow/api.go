package wow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"wowsim/models"

	"github.com/joho/godotenv"
)

//----------------------------- WoW Stuff -----------------------------

//to avoid generating several time the auth token for // connection
var mux sync.Mutex

//Store token per region
var authTokens map[string]AuthToken

//AuthToken simple object to restore password
type AuthToken struct {
	AccessToken string    `json:"access_token,omitempty"`
	TokenType   string    `json:"token_type,omitempty"`
	ExpiresIn   int       `json:"expires_in,omitempty"`
	Expiration  time.Time `json:"-"`
}

//IsValid check if auth token has expired
func (a *AuthToken) IsValid() bool {
	if a.Expiration.Before(time.Now()) {
		return false
	}
	return true
}

//Some stuff
const (
	urlEU = "eu.battle.net"
	urlUS = "us.battle.net"
	urlKR = "kr.battle.net"
	urlTW = "tw.battle.net"
	urlCN = "www.battlenet.com.cn"

	urlAPIEU = "eu.api.blizzard.com"
	urlAPIUS = "us.api.blizzard.com"
	urlAPIKR = "kr.api.blizzard.com"
	urlAPITW = "tw.api.blizzard.com"
	urlAPICN = "gateway.battlenet.com.cn"

	pathAuth                = "/oauth/token?grant_type=client_credentials"
	pathCharacterEquipment  = "/profile/wow/character/%s/%s/equipment?namespace=profile-%s"       //&locale=en_US
	pathCharacterAppearance = "/profile/wow/character/%s/%s/appearance?namespace=profile-%s"      //&locale=en_US
	pathCharacterMedia      = "/profile/wow/character/%s/%s/character-media?namespace=profile-%s" //&locale=en_US
	pathItemID              = "/data/wow/item/%s?namespace=static-%s"                             //&locale=en_US
	pathMediaItemID         = "/data/wow/media/item/%s?namespace=static-%s"                       //&locale=en_US
)

var clientID string
var clientSecret string

var regionToURL = map[string]string{
	"cn": urlAPICN,
	"eu": urlAPIEU,
	"us": urlAPIUS,
	"tw": urlAPITW,
	"kr": urlAPIKR,
}

var urlToRegion = map[string]string{
	urlAPICN: "cn",
	urlAPIEU: "eu",
	urlAPIUS: "us",
	urlAPIKR: "kr",
	urlAPITW: "tw",
	urlEU:    "eu",
	urlUS:    "us",
	urlCN:    "cn",
	urlKR:    "kr",
	urlTW:    "tw",
}

var urlAPItoAuth = map[string]string{
	urlAPICN: urlCN,
	urlAPIEU: urlEU,
	urlAPIUS: urlUS,
}

func auth(clientID, clientSecret, baseServer string) (AuthToken, error) {
	ret := AuthToken{}
	response, err := http.Post(fmt.Sprintf("https://%s:%s@%s%s", clientID, clientSecret, baseServer, pathAuth), "application/json", nil)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return ret, err
	}
	data, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(data, &ret)
	ret.Expiration = time.Now().Add(time.Second * time.Duration(ret.ExpiresIn))
	if err != nil {
		log.Printf("%+v\nErreur : %s\n", ret, err)

	}
	if authTokens == nil {
		authTokens = map[string]AuthToken{baseServer: ret}
	} else {
		authTokens[baseServer] = ret
	}
	return ret, nil
}

//return the auth if existing
func getAuthBearer(baseServer string) (AuthToken, error) {
	mux.Lock()
	defer mux.Unlock()
	if val, ok := authTokens[baseServer]; ok && val.IsValid() {
		log.Println("Using cached Token")
		return val, nil
	}
	if clientID == "" && clientSecret == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
		clientID = os.Getenv("CLIENT_ID")
		clientSecret = os.Getenv("CLIENT_SECRET")
	}
	log.Println("Query Token")
	return auth(clientID, clientSecret, baseServer)
}

//GetCharacterEquipment return caracter equipement
func GetCharacterEquipment(region, realm, name string) models.CharacterMeta {
	ret := models.CharacterMeta{}
	url := fmt.Sprintf(pathCharacterEquipment, realm, name, region)
	getAPIResponse(&ret, url, region)
	return ret
}

//GetItem return an item by it's is for a region
func GetItem(region, id string) models.Item {
	url := fmt.Sprintf(pathItemID, id, region)
	ret := models.Item{}
	getAPIResponse(&ret, url, region)
	return ret
}

//GetItemMedia return an item by it's is for a region
func GetItemMedia(region, id string) models.ItemMedia {
	url := fmt.Sprintf(pathMediaItemID, id, region)
	ret := models.ItemMedia{}
	getAPIResponse(&ret, url, region)
	return ret
}

//GetCharacterAppearance return /appearance
func GetCharacterAppearance(region, realm, name string) models.CharacterAppearance {
	url := fmt.Sprintf(pathCharacterAppearance, realm, name, region)
	ret := models.CharacterAppearance{}
	getAPIResponse(&ret, url, region)
	return ret
}

//GetCharacterMedia return /appearance
func GetCharacterMedia(region, realm, name string) models.CharacterMedia {
	url := fmt.Sprintf(pathCharacterMedia, realm, name, region)
	ret := models.CharacterMedia{}
	getAPIResponse(&ret, url, region)
	return ret
}

func getAPIResponse(item interface{}, url, region string) {
	baseServer := regionToURL[region]
	token, err := getAuthBearer(urlAPItoAuth[baseServer])
	if err != nil {
		log.Printf("%s\n", err.Error())
	}

	log.Println("Accessing API : ", url)

	client := &http.Client{}
	req, nil := http.NewRequest("GET", fmt.Sprintf("https://%s%s", baseServer, url), nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err = json.Unmarshal(data, item)
		if err != nil {
			log.Printf("%s\n", string(data))
			log.Printf(err.Error())
		}
	}
}

/*
jsonData := map[string]string{"firstname": "Nic", "lastname": "Raboy"}
jsonValue, _ := json.Marshal(jsonData)
response, err = http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(jsonValue))

func getTT() {
	fmt.Println("Starting the application...")
	response, err := http.Get("https://httpbin.org/ip")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
	jsonData := map[string]string{"firstname": "Nic", "lastname": "Raboy"}
	jsonValue, _ := json.Marshal(jsonData)
	response, err = http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
	fmt.Println("Terminating the application...")
}

*/
