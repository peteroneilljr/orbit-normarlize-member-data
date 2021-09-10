# Orbit Data Normalizer

Thanks for checking out my handy Orbit data normalizer tool. 

## Syncronize Member Location Data

1. Query member profiles for a given word and return either location or company data. This will search their user bio as well, so some results may match on `ex-apple` or similar.
    ```bash
    $ go run ./main.go --query="brisbane" --return-location | sort -u | paste -sd " " -
    "Brisbane" "Brisbane, Australia" "San Francisco"
    ```

1. Trim the unwanted results and search for matches.
    ```bash
    $ go run ./main.go --field=location "Brisbane" "Brisbane, Australia"
    Orbit Field: location
    New Name:
    Old Names: [Brisbane Brisbane, Australia]
    Checking for members with location: Brisbane
    Number of matching members: 1
    Checking for members with location: Brisbane, Australia
    Number of matching members: 12

    ===
    Number Of Members Updated: 0
    ```
1. Choose the value you wish to synchronize across these member accounts and specify it with the `--new` flag. 
    > WARNING: This change is NOT reversable. Once these changes are made they are permanent.
    ```bash
    $ go run ./main.go --field=location --new="Brisbane, Australia"  "Brisbane"
    Orbit Field: location
    New Name: Brisbane, Australia
    Old Names: [Brisbane]
    Checking for members with location: Brisbane
    Number of matching members: 1
    Member Name: Kangaroo Roo
    Member Email:
    Member ID: xxxxxx
    location changed to: Brisbane, Australia

    ===
    Number Of Members Updated: 1
    ```

1. Run the query again without the `--new` flag to see that all users have been moved over to the correct field.
    ```bash
    $ go run ./main.go --field=location "Brisbane, Australia"  "Brisbane"
    Orbit Field: location
    New Name:
    Old Names: [Brisbane, Australia Brisbane]
    Checking for members with location: Brisbane, Australia
    Number of matching members: 13
    Checking for members with location: Brisbane
    Number of matching members: 0

    ===
    Number Of Members Updated: 0
    ```

## Syncronize Member Company Data

1. Query member profiles for a given word and return either location or company data. This will search their user bio as well, so some results may match on `ex-apple` or similar.
    ```bash
    $ go run ./main.go --query="apple" --return-company | sort -u | paste -sd " " -
    "" "@apple" "@atlassian " "Apple Inc." "Apple" "Roku" "ThoughtWorks"
    ```
1. Trim the unwanted results and search for matches.
    ```bash
    $ go run ./main.go --field=company "@apple" "Apple Inc." "Apple"
    Orbit Field: company
    New Name:
    Old Names: [@apple Apple Inc. Apple]
    Checking for members with company: @apple
    Number of matching members: 2
    Checking for members with company: Apple Inc.
    Number of matching members: 1
    Checking for members with company: Apple
    Number of matching members: 3

    ===
    Number Of Members Updated: 0
    ```
1. Choose the value you wish to synchronize across these member accounts and specify it with the `--new` flag. 
    > WARNING: This change is NOT reversable. Once these changes are made they are permanent.
    ```bash
    $ go run ./main.go --field=company --new="Apple Inc." "@apple" "Apple"
    Orbit Field: company
    New Name: Apple Inc.
    Old Names: [@apple Apple]
    Checking for members with company: @apple
    Number of matching members: 2
    Member Name: Green Fish 
    Member Email:
    Member ID: xxxxxx
    company changed to: Apple Inc.
    Member Name: Red Hen
    Member Email:
    Member ID: xxxxxx
    company changed to: Apple Inc.
    Checking for members with company: Apple
    Number of matching members: 3
    Member Name: Blue Fish
    Member Email: Blue_fish@gmail.com
    Member ID: xxxxxx
    company changed to: Apple Inc.
    Member Name: Sandy Turtle
    Member Email:
    Member ID: xxxxxx
    company changed to: Apple Inc.
    Member Name: Billy Goat
    Member Email: Billy@Goat.com
    Member ID: xxxxxx
    company changed to: Apple Inc.

    ===
    Number Of Members Updated: 5
    ```


1. Run the query again without the `--new` flag to see that all users have been moved over to the correct field.
    ```bash
    $ go run ./main.go --field=company "Apple Inc." "@apple" "Apple"
    Orbit Field: company
    New Name:
    Old Names: [Apple Inc. @apple Apple]
    Checking for members with company: Apple Inc.
    Number of matching members: 6
    Checking for members with company: @apple
    Number of matching members: 0
    Checking for members with company: Apple
    Number of matching members: 0

    ===
    Number Of Members Updated: 0
    ```

## Query for user data

Query without either of the return flags just returns the raw json object from Orbit. This works really well for searching and pulling member attributes. 

```bash
$ go run ./main.go --query=apple | jq '.data[] | "\(.attributes.email) \(.attributes.github)"'
"null null"
"null bigfish"
"null green-apple"
"null null"
"user@gmail.com kitkat1919"
"mossgarden@gmail.com mgarden"
"opauser@gmail.com opauser"
"null scooterguy"
"null jackolanter"
"me@email.com null"
"big@ryhme.com bigryhme"
"null null"
```

## Parameters

- `--field`: This selects the member attribute to search, currently *location* or *company* are available
- Names: Pass in space-separated strings at the end of the command. Orbit will query the members that match these names.
- `--new`: The value to syncronize across members for the provided field
- `--query`: Search the member profiles for a given word or phrase.
- `--return-location`: Returns a list of location data for the matching members
- `--return-company`: Returns a list of company data for the matching members


- `ORBIT_API_KEY`: Set your Orbit API key as an environment variable
- `ORBIT_WORKSPACE_ID`: Set the Orbit workspace ID you want to modify