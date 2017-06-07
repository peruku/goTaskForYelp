
package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
	"bytes"
	"io/ioutil"
	"net/url"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	mClientId = "y2CWt70sH1j83IpesaWRFA"
	mClientSecret = "0zZOiHiY2Urocnp1me6PKlCUlV0nlm79kTbu0Jwd5y9Xyg9awqzxMSK4SmvOGl9K"
	mUrlAccessToken = "https://api.yelp.com/oauth2/token"
	mUrlBusinessSearch = "https://api.yelp.com/v3/businesses/search"
)

type ParamSearchBusiness struct
{
	Term string `json:"term"`
	Location string `json:"location"`
}

//For performance, return the required fields only
type RespSearchBusinessMin struct {
	Term 		string `json:"term"`
	Location 	string `json:"location"`
	Total      	int64      `json:"total"`
	Businesses 	[]BusinessMin `json:"businesses"`
}

//For performance, return the required fields only
type BusinessMin struct {
	Rating      float32 `json:"rating"`
	Name        string  `json:"name"`
	ImgUrl      string  `json:"image_url"`
	Location	struct {
					City 		string `json:"city"`
					Country 	string `json:"country"`
					Address2 	string `json:"address2"`
					Address3 	string `json:"address3"`
					State 		string `json:"state"`
					Address1 	string `json:"address1"`
					ZipCode 	string `json:"zip_code"`
	} `json:"location"`
	Coordinates  struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
	} `json:"coordinates"`
}

type RespSearchBusiness struct {
	Term 		string `json:"term"`
	Location 	string `json:"location"`
	Total      	int64  `json:"total"`
	Businesses 	[]Business `json:"businesses"`
	Region 		struct {
				Center struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
				} `json:"center"`
			} `json:"region"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Business struct {
	Rating 		float32 `json:"rating"`
	Price 		string  `json:"price"`
	Id 		string  `json:"id"`
	IsClosed 	bool `json:"is_closed"`
	Categories 	[]struct {
					Alias string `json:"alias"`
					Title string `json:"title"`
	} `json:"categories"`
	ReviewCount	int32 `json:"review_count"`
	Name 		string  `json:"name"`
	Url 		string  `json:"url"`
	Coordinates 	struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
	} `json:"coordinates"`
	ImgUrl 		string  `json:"image_url"`
	Location	struct {
					City 		string `json:"city"`
					Country 	string `json:"country"`
					Address2 	string `json:"address2"`
					Address3 	string `json:"address3"`
					State 		string `json:"state"`
					Address1 	string `json:"address1"`
					ZipCode 	string `json:"zip_code"`
	} `json:"location"`
	Distance	float32 `json:"distance"`
	Transactions	[]string `json:"transactions"`
}

//Get Yelp access token
func GetAccessToken(pAccessToken *AccessToken) error {
	lUrlValues := url.Values{}
	lUrlValues.Add("grant_type", "client_credentials")
	lUrlValues.Add("client_id", mClientId)
	lUrlValues.Add("client_secret", mClientSecret)

	lBufferUrlValues := bytes.NewBuffer([]byte(lUrlValues.Encode()))

	lResponse, lErr := http.Post(mUrlAccessToken, "application/x-www-form-urlencoded", lBufferUrlValues)
	if lErr != nil {
		return lErr
	}

	defer lResponse.Body.Close()
	lByteArr, lErr := ioutil.ReadAll(lResponse.Body)
	if lErr != nil {
		return lErr
	}

	lErr = json.Unmarshal(lByteArr, &pAccessToken)
	if lErr != nil {
		return lErr
	}

	return nil
}

//Yelp business search
func SearchBusinessMin(pParamSearchBusiness ParamSearchBusiness, pAccessTokenPtr *AccessToken) (*RespSearchBusinessMin, error) { 
	lUrlWithParam := mUrlBusinessSearch + "?" + "term=" + pParamSearchBusiness.Term + "&location=" + pParamSearchBusiness.Location

	lRequest, lErr := http.NewRequest("GET", lUrlWithParam, nil)
	if lErr != nil {
		return nil, lErr
	}

	lRequest.Header.Set("Authorization", pAccessTokenPtr.TokenType + " " + pAccessTokenPtr.AccessToken)

	lRefHttpClient := &http.Client{}
	lResponse, lErr := lRefHttpClient.Do(lRequest)
	if lErr != nil {
		return nil, lErr
	}

	lByteArr, lErr := ioutil.ReadAll(lResponse.Body)
	if lErr != nil {
		return nil, lErr
	}

	var lRespSearchBusinessMin = new(RespSearchBusinessMin)
	lErr = json.Unmarshal([]byte(lByteArr), &lRespSearchBusinessMin)
	if lErr != nil {
		return nil, lErr
	}
	
	lRespSearchBusinessMin.Term = pParamSearchBusiness.Term
	lRespSearchBusinessMin.Location = pParamSearchBusiness.Location

	return lRespSearchBusinessMin, nil
}

//Yelp business search
func SearchBusiness(pParamSearchBusiness ParamSearchBusiness, pAccessTokenPtr *AccessToken) (*RespSearchBusiness, error) { 
	lUrlWithParam := mUrlBusinessSearch + "?" + "term=" + pParamSearchBusiness.Term + "&location=" + pParamSearchBusiness.Location

	lRequest, lErr := http.NewRequest("GET", lUrlWithParam, nil)
	if lErr != nil {
		return nil, lErr
	}

	lRequest.Header.Set("Authorization", pAccessTokenPtr.TokenType + " " + pAccessTokenPtr.AccessToken)

	lRefHttpClient := &http.Client{}
	lResponse, lErr := lRefHttpClient.Do(lRequest)
	if lErr != nil {
		return nil, lErr
	}

	lByteArr, lErr := ioutil.ReadAll(lResponse.Body)
	if lErr != nil {
		return nil, lErr
	}

	var lRespSearchBusiness = new(RespSearchBusiness)
	lErr = json.Unmarshal([]byte(lByteArr), &lRespSearchBusiness)
	if lErr != nil {
		return nil, lErr
	}

	return lRespSearchBusiness, nil
}


const (
	mFolderLibs string = "libs"
	mFolderHtml string = "html"
)

var (
	mYelpAccessTokenTimeExpired time.Time 
	mYelpAccessToken *AccessToken 
)

//Render Html pages
func RenderPage(pResponseWriter http.ResponseWriter, pHtmlFileName string) {
	log.Println("Processing " + pHtmlFileName + ".html... " )
	lListHtmlFile := []string{ mFolderHtml + "/base.html", 
				mFolderHtml + "/" + pHtmlFileName + ".html" }

	lTemplatePtr, lErr := template.ParseFiles(lListHtmlFile...)
	if lErr != nil {
		log.Println("Error parsing files: ", lErr)
	}

	lErr = lTemplatePtr.Execute(pResponseWriter, nil)
	if lErr != nil {
		log.Println("Error executing template: ", lErr)
	}
}

//Handler for processing files in libs folder
func HandlerFilesLib(pResponseWriter http.ResponseWriter, pHttpRequest *http.Request) {
	lPathLibFile := pHttpRequest.URL.Path[len("/" + mFolderLibs + "/"):]
	
	if len(lPathLibFile) != 0 {
		lFile, lErr := http.Dir(mFolderLibs + "/").Open(lPathLibFile)
		if lErr != nil {
			log.Println("Error opening lib file: ", lErr)
			http.NotFound(pResponseWriter, pHttpRequest)
		} else {
			lReadSeeker := io.ReadSeeker(lFile)
			http.ServeContent(pResponseWriter, pHttpRequest, lPathLibFile, time.Now(), lReadSeeker)
		}
	}
}
//mongo db related
func CreateRespSearchBusinessMin(pRespSearchBusinessMin  *RespSearchBusinessMin) (bool, error) {
	lSession, lErr := mgo.Dial("localhost:27017")
	if lErr != nil {
		return false, lErr
	}
	defer lSession.Close()
    // Database name restApp5 and Collection name is RespSearchBusinessMin
	lCollection := lSession.DB("restApp5").C("RespSearchBusinessMin")
	lErr = lCollection.Insert(pRespSearchBusinessMin)
	if lErr != nil {
		return false, lErr
	}
	return true, lErr
}

func GetRespSearchBusinessMin(pTerm string, pLocation string) (*RespSearchBusinessMin, error) {
	var lRespSearchBusinessMin = new(RespSearchBusinessMin)

	lSession, lErr := mgo.Dial("localhost:27017")
	if lErr != nil {
		return lRespSearchBusinessMin, lErr
	}
	defer lSession.Close()

	lCollection := lSession.DB("restApp5").C("RespSearchBusinessMin")
	lErr = lCollection.Find(bson.M{"term": pTerm, "location": pLocation}).One(&lRespSearchBusinessMin)
	if lErr != nil {
		return lRespSearchBusinessMin, lErr
	}
	return lRespSearchBusinessMin, lErr
}

//Handler for processing html/index.html
func HandlerPageIndex(pResponseWriter http.ResponseWriter, pHttpRequest *http.Request) {
	RenderPage(pResponseWriter, "index")
}

//Handler for searching business in yelp
func HandlerSearchBusinessInYelp(pResponseWriter http.ResponseWriter, pRequest *http.Request) {
	pResponseWriter.Header().Set("Content-Type", "application/json")

	var lYelpParamSearchBusiness ParamSearchBusiness 
	lDecoder := json.NewDecoder(pRequest.Body)
	lErr := lDecoder.Decode(&lYelpParamSearchBusiness)
	if lErr != nil {
	        log.Println("Error decoding Json string: ", lErr)
	}
	defer pRequest.Body.Close()


	lYelpRespSearchBusinessMinFromMongoDB, lErr := GetRespSearchBusinessMin(lYelpParamSearchBusiness.Term, lYelpParamSearchBusiness.Location)
	if lErr == nil {
		if lYelpRespSearchBusinessMinFromMongoDB.Term == lYelpParamSearchBusiness.Term {
			json.NewEncoder(pResponseWriter).Encode(lYelpRespSearchBusinessMinFromMongoDB)
		        log.Println("Retrieved data from Mongo DB...")
			return
		}
	}

	if IsYelpAccessTokenExpired() {
		SetYelpAccessToken()
	}

	log.Println("Yelp access token expiring on " + mYelpAccessTokenTimeExpired.String() + "...") 

	lYelpRespSearchBusinessMin, lErr := SearchBusinessMin(lYelpParamSearchBusiness, mYelpAccessToken)
    	if lErr != nil {
	        log.Println("Error searching businesses in Yelp: ", lErr)
    	}

	json.NewEncoder(pResponseWriter).Encode(lYelpRespSearchBusinessMin)

	lInsertIntoMongo, lErr := CreateRespSearchBusinessMin(lYelpRespSearchBusinessMin)
	if lErr != nil {
	        log.Println("Error decoding Json string: ", lErr)
	}

	log.Println("Insert into Mongo: ", lInsertIntoMongo)
}

func IsYelpAccessTokenExpired() (bool) {
	if time.Now().After(mYelpAccessTokenTimeExpired) {
		log.Println("Yelp access token expired on " + mYelpAccessTokenTimeExpired.String()) 
		return true
	}
	return false
}

func SetYelpAccessToken() {
        log.Println("Retrieving Yelp access token...")
	mYelpAccessTokenTimeExpired = time.Now()
	mYelpAccessToken = new(AccessToken)
	lErr := GetAccessToken(mYelpAccessToken)
	if lErr != nil {
		log.Println("Error retrieving Yelp access token: ", lErr)
	}
	lDuration := time.Duration(mYelpAccessToken.ExpiresIn*1000*1000*1000)
	mYelpAccessTokenTimeExpired = mYelpAccessTokenTimeExpired.Add(lDuration)
}

//Main function
func main(){
	SetYelpAccessToken()

        log.Println("Registering handlers...")
	http.HandleFunc("/" + mFolderLibs + "/", HandlerFilesLib)
	http.HandleFunc("/", HandlerPageIndex)
	http.HandleFunc("/SearchBusiness", HandlerSearchBusinessInYelp)
	
        log.Println("Starting webserver...")
	lErrListenAndServe := http.ListenAndServe(":8000", nil)
	if lErrListenAndServe != nil {
	        log.Println("Error starting webserver: ", lErrListenAndServe)
	}
}