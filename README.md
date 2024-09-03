# Go Bootcamp Project - Day 03

This project was developed as part of the Go Bootcamp, focusing on creating a simple API with data loading, getting data from store(interface hidden elasticsearch), simple auth and find places using elasticsearch engine.
In project parse data from data.csv about eating places at Moscow and push this data using go-routines to elasticsearch.

## Overview

This project includes:

- **Data Loading:** A module that handles the loading of initial data required by the API.
- **Simple Interface:** To inverse requerements to store.
- **Complete API:** A fully functional REST API built with Go Mux.
- **Auth** Using JWT bearer.
- **Find place to eat by coords** Using elasticsearch sort options.
- **Paginate by page size** To pretty view and reduce the load.

## Features

- **Modular Architecture:** Modular code organization following best practices.
- **Dockerized:** Easily deployable using Docker. Simply run `docker-compose up` to start the application.
- **RESTful API:** Implements a basic but fully functional REST API with standard HTTP methods.
- **Swagger Docs** Implemented documentation generation using swagger 2.0 and openAPI. at `/swagger/`
- **Auth** for secure endpoint `/api/recommend/`.

 <table>
  <tr>
    <th>Routes</th>
  </tr>
  <tr>
    <td>
        <img width="1469" alt="image" src="https://github.com/user-attachments/assets/764d6655-d838-4ef9-9b1b-647bfc88056a">
    </td>
  </tr>
</table> 



## Getting Started



### Prerequisites

- Docker
- Docker Compose

### Installation

1. Clone this repository:
    ```bash
    git clone https://github.com/zkhrg/go_day03.git
    ```

2. Start the project using Docker Compose:
    ```bash
    cd go_day03 && docker compose up
    ```

3. The API will be available at `http://localhost:8888`.

## Usage

The API provides endpoints to interact with the data loaded into the system. You can test the API using tools like Postman or `curl`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
