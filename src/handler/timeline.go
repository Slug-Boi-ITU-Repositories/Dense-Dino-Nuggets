package handler
/*
import (
    "log"
    "net/http"
    "text/template"
    "minitwit/src/repository"
    // ... other imports
)

const PER_PAGE = 30

var (
    messageRepo = &repository.MessageRepository{}
    userRepo    = &repository.UserRepository{}
)

func Timeline(w http.ResponseWriter, r *http.Request) {
    user, err := getUser(r)
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if user == nil {
        http.Redirect(w, r, "/public", http.StatusFound)
        return
    }

    messages, err := messageRepo.GetPersonalTimeline(user.UserID, PER_PAGE)
    if err != nil {
        log.Println(err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // ... rest of your template rendering
}*/