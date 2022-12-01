# Workout Logger API
<img width="1440" alt="Screen Shot 2022-11-13 at 11 47 42 AM" src="https://user-images.githubusercontent.com/46465568/201538851-99b051a7-b084-4919-993c-93f7efffd447.png">


The GraphQL api for a very simple workout logger app I built for myself to track progression of weight I'm lifting for certain exercises.

# Try It Out
1. Go to https://workout-logger-api-ejtky726bq-uw.a.run.app/
2. Do a login mutation to get the access token by pasting and running 
```
mutation Login {
  login(email: "test@test.com", password:"password123") {
    ... on AuthSuccess {
      accessToken
      refreshToken
    }
  }
}
```
3. take the access token and paste this into the header section at the bottom of the page 
```
{
  Authorization:"Bearer <PASTE_ACCESS_TOKEN_HERE>"
}
```
4. Click the docs button on the top left of the page and start running queries!

# Techonologies Used

- Go
- GORM
- GQLGen
- GraphQL
- PostgreSQL

# Prereqs

- A postgres db url
- Go installed on your machine

# Setup

1. Clone repo
2. `cd` into the root of the repo
3. Have copy contents of `.test.env` into a new `.env` file
4. Fill in and replace secrets and postgres database connection parameters 
5. Run `make dev` to start dev server or `make test` to run all integration tests

# Commands

- `make dev`: start dev environment
- `make test`: run all test files
- `make format`: format all code within repo
- `make regenerate`: regenerate graphql resolvers from `schema.graphqls`
