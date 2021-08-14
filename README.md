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
     "email": "tony@stark.com"
   }
   ```