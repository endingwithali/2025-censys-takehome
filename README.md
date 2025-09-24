# Censys Take Home 2025

### Backend
- **Language:** Go (Golang) 1.24+
- **Frameworks/Libraries:** 
  - [chi](https://github.com/go-chi/chi) (HTTP router)
  - [gorm](https://gorm.io/) (ORM for PostgreSQL)
  - [github.com/go-chi/cors](https://github.com/go-chi/cors) (CORS middleware)
  - [github.com/nsf/jsondiff](https://github.com/nsf/jsondiff) (difference finder)
- **Database:** PostgreSQL

### Frontend
- **Language:** JavaScript (React)
- **Frameworks/Libraries:**
  - [React](https://react.dev/) (UI library)
  - [axios](https://axios-http.com/) (HTTP client)
  - [react-dropzone](https://react-dropzone.js.org/) (file upload)
  - [date-fns](https://date-fns.org/) (date formatting)

Instructions on how to run each part are located in the `README.md` files of `/backend` and `/frontend`. 

## Testing
Automated testing was written for the backend. Frontend was tested manually. 


## AI Techniques 
I utilized AI to build the frontend, and used AI to help create the scaffolding for the backend code. I used Claude to write the frontend, and many of the tests for the backend. I asked Claude to help clean up the tests I did write for the backend to fit into its testing suite. I used ChatGPT to help with development of backend code to make understanding certain Golang APIs faster. 

## Assumptions and Explanations
### Assumptions:
1) Files are less than 25 MiB 
2) Comparisons only needed to be made between two different timestamps of a specific host 

## Explanations:
My original implementation involved creating a database for storing already created differences. Given the scope of the work, I decided not to move forward with the recording of calculated differences. In practice, for larger files and systems, it would be advantageous to store differences to reduce calculation loads on the server (especially for larger files) since once calculated they will never change. I commented out the beginning of the work for this, however decided to keep it in to demonstrate how this project can be continued and grown without major reworks of the already completed work.
(`/backend/internal/service/differences_service.go`, `/backend/internal/repo/difference.go`, `/backend/internal/repo/schema/schema.sql`)

In code, I did include comments on some normally proactive actions I chose to bypass given the scope of work, including checking to see if a comparison is being attempted using the same file. I chose not to use `.env` for the frontend and backend systems given the scope of the work.

## Future Enhancements
Future enhancements include saving the precalculated differences, containerization to allow for easier cross machine deployment, adding `.env` support, adding cross-host snapshot difference assessment capability, and more rigorous testing of both the backend and frontend. 
