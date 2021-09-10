package main

import (
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
var validSearch = true
var newName string
var oldNames []string
var numMembersChanged = 0

func validateRequest() {
	if orbitApiKey == "" {
		fmt.Println("Please provided an API key using env var ORBIT_API_KEY")
		validSearch = false
	}
	if orbitWorkspaceID == "" {
		fmt.Println("Please provided your Orbit Workspace ID using env var ORBIT_WORKSPACE_ID")
		validSearch = false
	}

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
func updateMemberList(memberList []byte, memberCount int, field string, name string) {
	for x := 0; x < memberCount; x++ {

		memberName := gjson.GetBytes(memberList, fmt.Sprintf("data.%d.attributes.name", x))
		fmt.Println("Member Name:", memberName)

		memberEmail := gjson.GetBytes(memberList, fmt.Sprintf("data.%d.attributes.email", x))
		fmt.Println("Member Email:", memberEmail)

		memberID := gjson.GetBytes(memberList, fmt.Sprintf("data.%d.attributes.id", x))
		fmt.Println("Member ID:", memberID)

		updateMember(memberID.String(), field, name)
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
				updateMemberList(matchingMembers, matchingCount, orbitField, newName)
			}
		}
		fmt.Printf("\n===\nNumber Of Members Updated: %d\n", numMembersChanged)
	}

}
