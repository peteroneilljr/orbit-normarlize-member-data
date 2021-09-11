package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

var orbitApiKey = os.Getenv("ORBIT_API_KEY")
var orbitWorkspaceID = os.Getenv("ORBIT_WORKSPACE_ID")
var orbitField string
var orbitQuery string
var returnLocation bool
var returnCompany bool
var newName string
var oldNames []string
var numMembersChanged = 0

type Member struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Title     string   `json:"title"`
	Avatar    string   `json:"avatar_url"`
	Bio       string   `json:"bio"`
	Birthday  string   `json:"birthday"`
	Company   string   `json:"company"`
	Location  string   `json:"location"`
	Tags      []string `json:"tags"`
	Teammate  bool     `json:"teammate"`
	Url       string   `json:"url"`
	OrbitUrl  string   `json:"orbit_url"`
	Twitter   string   `json:"twitter"`
	GitHub    string   `json:"github"`
	Discourse string   `json:"discourse"`
	Discord   string   `json:"discord"`
	DevTo     string   `json:"devto"`
	Linkedin  string   `json:"linkedin"`
}

func validateRequest() {
	var validSearch = true

	// Check for valid Orbit API credentials
	if orbitApiKey == "" {
		fmt.Println("Please provided an API key using env var ORBIT_API_KEY")
		validSearch = false
	}
	if orbitWorkspaceID == "" {
		fmt.Println("Please provided your Orbit Workspace ID using env var ORBIT_WORKSPACE_ID")
		validSearch = false
	}

	// Check for update parameters
	if orbitQuery == "" {
		if orbitField == "" {
			fmt.Println("Please provided the field you wish to scan with: --field")
			validSearch = false
		}
		if len(oldNames) == 0 {
			fmt.Println("Please provide the names you want to search for after command line flags")
			validSearch = false
		}
	} else {
		// Check that no flags are passed in with query flag
		if orbitField != "" {
			fmt.Println("--query is not compatible with --field")
			validSearch = false
		}
		if newName != "" {
			fmt.Println("--query is not compatible with --new")
			validSearch = false
		}
		if len(oldNames) != 0 {
			fmt.Println("--query is not compatible with arguments after the command line")
			validSearch = false
		}
	}
	if !validSearch {
		os.Exit(1)
	}
}

func updateMember(memberID string, field string, name string) {

	url := fmt.Sprintf("https://app.orbit.love/api/v1/%s/members/%s", orbitWorkspaceID, memberID)

	payloadString := fmt.Sprintf("{\"%s\":\"%s\"}", field, name)
	payload := strings.NewReader(payloadString)

	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", orbitApiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 204 {
		fmt.Printf("%s changed to: %s\n", field, name)
		numMembersChanged++
	} else {
		fmt.Println("Error: HTTP Status Code:", res.StatusCode)
	}
}
func updateMemberList(memberList []byte, field string, name string) {
	membersIDs := gjson.GetBytes(memberList, "data.#.attributes.id")
	memberCount := len(membersIDs.Array())

	for x := 0; x < memberCount; x++ {

		member := Member{}
		memberData := gjson.GetBytes(memberList, fmt.Sprintf("data.%d.attributes", x))
		err := json.Unmarshal([]byte(memberData.Raw), &member)
		if err != nil {
			fmt.Println("Could not find info in:", memberList)
		}
		fmt.Println("Member Name:", member.Name)
		fmt.Println("Member Email:", member.Email)
		fmt.Println("Member ID:", member.Id)

		updateMember(member.Id, field, name)
	}
}
func getMemberList(field string, search string) []byte {
	url := fmt.Sprintf("https://app.orbit.love/api/v1/%s/members?%s=%s", orbitWorkspaceID, field, url.PathEscape(search))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", orbitApiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	memberListJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	if res.StatusCode != 200 {
		fmt.Println("Error finding members: HTTP Status Code:", res.StatusCode)
		os.Exit(1)
	}

	return memberListJson
}

func printMemberData(json []byte, field string, query string) {
	searchString := fmt.Sprintf("data.#.attributes.%s", field)
	results := gjson.GetBytes(json, searchString)

	if results.Raw == "[]" {
		fmt.Printf("No members for query: %s\n", query)
		return
	}

	results.ForEach(func(key, value gjson.Result) bool {
		fmt.Printf("\"%s\"\n", value.String())
		return true // keep iterating
	})

}

func main() {

	flag.StringVar(&orbitField, "field", "", "The field in Orbit you wish to update")
	flag.StringVar(&newName, "new", "", "This will replace the old data")
	flag.StringVar(&orbitQuery, "query", "", "This will return a list of members profiles that contain the query string")
	flag.BoolVar(&returnLocation, "return-location", false, "Returns a list of member locations with --query")
	flag.BoolVar(&returnCompany, "return-company", false, "Returns a list of member companies with --query")

	flag.Parse()
	oldNames = flag.Args()

	// Check that the requested fields are valid, if not exit program
	validateRequest()

	if orbitQuery != "" {
		memberList := getMemberList("query", orbitQuery)

		if returnCompany {
			printMemberData(memberList, "company", orbitQuery)
		}
		if returnLocation {
			printMemberData(memberList, "location", orbitQuery)
		}
		if !returnCompany && !returnLocation {
			fmt.Println(string(memberList))
		}

	} else if orbitField != "" {
		fmt.Println("Orbit Field:", orbitField)
		fmt.Println("New Name:", newName)
		fmt.Println("Old Names:", oldNames)
		for i := 0; i < len(oldNames); i++ {

			fmt.Printf("Checking for members with %s: %s\n", orbitField, oldNames[i])

			matchingMembers := getMemberList(orbitField, oldNames[i])
			memberIDs := gjson.GetBytes(matchingMembers, "data.#.attributes.id")
			matchingCount := len(memberIDs.Array())

			fmt.Println("Number of matching members:", matchingCount)

			if matchingCount == 0 {
				continue
			} else if newName != "" {
				updateMemberList(matchingMembers, orbitField, newName)
			}
		}
		fmt.Printf("\n===\nNumber Of Members Updated: %d\n", numMembersChanged)
	}

}
