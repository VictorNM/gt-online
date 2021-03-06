# GT ONLINE

## Prerequisites

You will need to have the following installed:

- Docker
- Make

## Usage

1. Clone the repo
   ```sh
   git clone https://github.com/VictorNM/gt-online.git
   ```
2. Run the app
   ```sh
   make up
   ```

3. In update code, you need to rebuild the app
    ```sh
   make build
   ```

4. Check if the app run successfully
    ```sh
   make log ## Should see "Server start at..."
   ```

5. For more information
   ```sh
   make
   ```

## API

All APIs will follow the below rules:

- Content-Type: "application/json"
- Authenticate: using header `Authorization`: `<token_type> <access_token>`
  - Example: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9ueUBzdGFyay5jb20ifQ.QfCpJBCrw4RzWM3OyDwiuTrZLAMefrSBF-YuVvodZoY`
- Error response: if the response HTTP status is not 2xx, an error will be return
   ```
   code:    string
   message: string
   ```
  Example:
    ```json
    {
     "code": "ALREADY_EXISTS",
     "message": "Email already registered."
    }
    ```

### Register

#### Request

- Method: POST
- Path: /auth/register
- Body:
   ```
   email:                  string, required
   password:               string, required
   password_confirmation:  string, required
   last_name:              string, required
   first_name:             string, required
   ```
  Example:
   ```json
   {
     "email": "tony@stark.com",
     "password": "abc@123@XYZ",
     "password_confirmation": "abc@123@XYZ",
     "last_name": "Stark",
     "first_name": "Tony"
   }
   ```

#### Response

- 200: Success
   ```json
   {
     "email": "tony@stark.com",
     "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9ueUBzdGFyay5jb20ifQ.QfCpJBCrw4RzWM3OyDwiuTrZLAMefrSBF-YuVvodZoY",
     "token_type": "Bearer"
   }
   ```

### Login

#### Request

- Method: POST
- Path: /auth/login
- Body:
   ```
   email:                  string, required
   password:               string, required
   ```
  Example:
   ```json
   {
     "email": "tony@stark.com",
     "password": "abc@123@XYZ"
   }
   ```

#### Response

- 200: Success
   ```json
   {
     "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidG9ueUBzdGFyay5jb20ifQ.QfCpJBCrw4RzWM3OyDwiuTrZLAMefrSBF-YuVvodZoY",
     "token_type": "Bearer"
   }
   ```

### List Users

#### Request

- Method: GET
- Path: /users
- Authenticate: yes
- Query (must provide at least 1 of these queries):
    ```
    email:    string
    name:     string
    hometown: string
    ```
  Example:
    ```
    /users?hometown=Metropolis&name=Tony"
    ```

#### Response

- 200: Success
    ```json
    {
      "count": 1,
      "users": [
        {
          "email": "tony@stark.com",
          "first_name": "Tony",
          "last_name": "Stark",
          "hometown": "New York"
        }
      ]
    }
    ```

### Get Profile

#### Request

- Method: GET
- Path: /users/profile
- Authenticate: yes

#### Response

- 200: Success
    ```json
    {
      "email": "tony@stark.com",
      "first_name": "Tony",
      "last_name": "Stark",
      "sex": "M",
      "birthdate": "29/05/1970",
      "current_city": "New York",
      "hometown": "New York",
      "interests": ["Technology"],
      "education": [
        {
            "school": "Harvard University",
            "year_graduated": 1992
        }
      ],
      "professional": [
        {
            "employer": "Alphabet",
            "job_title": "President"
        }
      ]
    }
    ```

### Update Profile

#### Request

- Method: PUT
- Path: /users/profile
- Authenticate: yes
- Body:
    ```
    email:                  string, required
    sex:                    string, enum: "M", "F"
    birthdate:              string, format: DD/MM/YYYY
    current_city            string
    hometown                string
    interests               []string
    education               []object
        - school            string
        - year_graduated    int
    professional            []object
        - employer          string
        - job_title         int
    ```
  Example:
    ```json
    {
      "sex": "M",
      "birthdate": "29/05/1970",
      "current_city": "New York",
      "hometown": "New York",
      "interests": ["Technology"],
      "education": [
        {
            "school": "Harvard University",
            "year_graduated": 1992
        }
      ],
      "professional": [
        {
            "employer": "Alphabet",
            "job_title": "President"
        }
      ]
    }
    ```

#### Response

- 200: Success
    ```json
    {
      "email": "tony@stark.com",
      "first_name": "Tony",
      "last_name": "Stark",
      "sex": "M",
      "birthdate": "29/05/1970",
      "current_city": "New York",
      "hometown": "New York",
      "interests": ["Technology"],
      "education": [
        {
            "school": "Harvard University",
            "year_graduated": 1992
        }
      ],
      "professional": [
        {
            "employer": "Alphabet",
            "job_title": "President"
        }
      ]
    }
    ```

### List Schools

#### Request

- Method: GET
- Path: /schools
- Authenticate: yes

#### Response

- 200: Success
   ```json
   {
     "schools": [
        {
          "schools_name": "Aukamm Elementary School",
          "type": "Elementary School"
        },
        {
          "schools_name": "Harvard University",
          "type": "University"
        }
     ]
   }
   ```

### List Employers

#### Request

- Method: GET
- Path: /employers
- Authenticate: yes

#### Response

- 200: Success
   ```json
   {
     "employers": [
        {
          "employers_name": "Microsoft"
        },
        {
          "employers_name": "Alphabet"
        }
     ]
   }
   ```

### List Friend Requests

#### Request

- Method: GET
- Path: /friends/requests
- Authenticate: yes

#### Response

- 200: Success
   ```json
   {
     "request_from": [
        {
          "email": "tony@stark.com",
          "relationship": "Teammate"
        }
     ],
     "request_to": [
        {
          "email": "steve.rogers@avengers.com",
          "relationship": "Teammate"
        }
     ]
   }
   ```
  
### Create Friend Request

#### Request

- Method: PUT
- Path: /friends/requests/:friend_email
  ```
  friend_email      string,required
  ```
- Authenticate: yes
- Body:
  ```
  relationship     string
  ```

#### Response

- 200: Success

### Accept Friend Request

#### Request

- Method: PUT
- Path: /friends/:friend_email
  ```
  friend_email      string,required
  ```
- Authenticate: yes

#### Response

- 200: Success

### List Friends

#### Request

- Method: GET
- Path: /friends
- Authenticate: yes

#### Response

- 200: Success
   ```json
   {
     "friends": [
        {
          "friend_email": "tony@stark.com",
          "relationship": "Teammate",
          "date_connected": "November 23, 2020"
        }
     ]
   }
   ```

### Delete Friend Request

#### Request

- Method: DELETE
- Path: /friends/:friend_email
  ```
  friend_email:      string,required
  ```
- Authenticate: yes
- Query
  ```
  action:  string, enum: "cancel", "reject"
  ```

#### Response

- 200: Success