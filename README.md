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