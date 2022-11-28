package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/syscrypt/boring-calendar/internal/model"
	"github.com/syscrypt/boring-calendar/internal/util"
)

const (
	ERR_CREATE_CALENDER = "error while creating new calendar entry"
)

type CalendarService struct {
	stmtCreateCalendar      *sql.Stmt
	stmtUserExists          *sql.Stmt
	stmtCreateUser          *sql.Stmt
	stmtCalendarEntryExists *sql.Stmt
	stmtDeleteEvents        *sql.Stmt
	log                     *logrus.Logger
	db                      *sql.DB
}

func NewCalendarService(db *sql.DB, ctx context.Context, log *logrus.Logger) (*CalendarService, error) {
	createQuery := "insert into calendar values(?, ?, ?)"
	ctxCreateCalendar, err := db.PrepareContext(ctx, createQuery)
	if err != nil {
		return nil, err
	}

	userExistsQuery := "select * from user where token = ?"
	ctxUserExists, err := db.PrepareContext(ctx, userExistsQuery)
	if err != nil {
		return nil, err
	}

	calendarEntryExistsQuery := "select * from calendar where user_token = ? and date = ?"
	ctxCalendarEntryExists, err := db.PrepareContext(ctx, calendarEntryExistsQuery)
	if err != nil {
		return nil, err
	}

	createUserQuery := "insert into user values(?)"
	ctxCreateUser, err := db.PrepareContext(ctx, createUserQuery)
	if err != nil {
		return nil, err
	}

	deleteEventsQuery := "delete from events where calendar_uuid = ?"
	ctxDeleteEvents, err := db.PrepareContext(ctx, deleteEventsQuery)
	if err != nil {
		return nil, err
	}

	return &CalendarService{
		log:                     log,
		db:                      db,
		stmtCreateCalendar:      ctxCreateCalendar,
		stmtUserExists:          ctxUserExists,
		stmtCreateUser:          ctxCreateUser,
		stmtCalendarEntryExists: ctxCalendarEntryExists,
		stmtDeleteEvents:        ctxDeleteEvents,
	}, nil
}

func (s *CalendarService) Close() {
	s.stmtCalendarEntryExists.Close()
	s.stmtCreateCalendar.Close()
	s.stmtCreateUser.Close()
	s.stmtDeleteEvents.Close()
	s.stmtUserExists.Close()
}

func (s *CalendarService) UserExists(userToken string, ctx context.Context) (bool, error) {
	cnt := 0

	rows, err := s.stmtUserExists.QueryContext(ctx, userToken)
	if err != nil {
		return false, err
	}
	for rows.Next() {
		cnt++
	}

	return cnt != 0, nil
}

func (s *CalendarService) CalendarEntryExists(userToken string, date int64, ctx context.Context) (string, error) {
	cnt := 0
	rows, err := s.stmtCalendarEntryExists.QueryContext(ctx, userToken, date)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	cal := &model.Calendar{
		Uuid:      "",
		UserToken: "",
		Date:      int64(0),
	}

	for rows.Next() {
		cnt++
		err = rows.Scan(&cal.Uuid, &cal.UserToken, &cal.Date)
		if err != nil {
			return "", err
		}
		break
	}

	if cal.Uuid == "" && cnt > 0 {
		return "", errors.New("no valid columns found")
	}

	return cal.Uuid, nil
}

func (s *CalendarService) DeleteEvents(calendarUuid string, ctx context.Context) error {
	_, err := s.stmtDeleteEvents.ExecContext(ctx, calendarUuid)
	return err
}

func (s *CalendarService) CreateEvents(events []*model.Event, calendarUuid string, userToken string, ctx context.Context) error {
	if len(events) == 0 {
		return nil
	}
	values := []interface{}{}

	createEventsQuery := "insert into events values"
	for _, e := range events {
		createEventsQuery += " (?, ?, ?, ?, ?),"
		values = append(values, uuid.NewString(), calendarUuid, e.Title, e.StartTime, e.EndTime)
	}
	createEventsQuery = createEventsQuery[:len(createEventsQuery)-1]

	ctxCreateEvents, err := s.db.PrepareContext(ctx, createEventsQuery)
	if err != nil {
		return err
	}

	_, err = ctxCreateEvents.ExecContext(ctx, values...)
	return err
}

func (s *CalendarService) CreateUser(userToken string, ctx context.Context) error {
	exists, err := s.UserExists(userToken, ctx)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("user already exists")
	}

	_, err = s.stmtCreateUser.ExecContext(ctx, userToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *CalendarService) Create(w http.ResponseWriter, r *http.Request) {
	calendar := &model.Calendar{}
	err := json.NewDecoder(r.Body).Decode(calendar)
	if err != nil {
		util.WriteInternalServer(w, err, ERR_CREATE_CALENDER, s.log)
		return
	}

	currentDay := (int64(time.Now().Unix()/60/60/24) * (60 * 60 * 24))
	calendar.Date = int64(calendar.Date/60/60/24) * (60 * 60 * 24)

	// omit if events are before the current day
	if calendar.Date < currentDay {
		util.WriteBadRequest(w, errors.New("day is before current day"), ERR_CREATE_CALENDER, s.log)
		return
	}

	if exists, err := s.UserExists(calendar.UserToken, r.Context()); err == nil && !exists {
		util.WriteBadRequest(w, errors.New("invalid user token"), ERR_CREATE_CALENDER, s.log)
		return
	} else if err != nil {
		util.WriteInternalServer(w, errors.New("invalid user token"), ERR_CREATE_CALENDER, s.log)
		return
	}

	calUuid, err := s.CalendarEntryExists(calendar.UserToken, calendar.Date, r.Context())
	if err != nil {
		util.WriteInternalServer(w, err, ERR_CREATE_CALENDER, s.log)
		return
	}

	if calUuid == "" {
		calendar.Uuid = uuid.NewString()
		_, err = s.stmtCreateCalendar.ExecContext(r.Context(), calendar.Uuid, calendar.UserToken, calendar.Date)
		if err != nil {
			util.WriteBadRequest(w, errors.New("day is before current day"), ERR_CREATE_CALENDER, s.log)
			return
		}
	} else {
		err = s.DeleteEvents(calUuid, r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			util.WriteBadRequest(w, err, ERR_CREATE_CALENDER, s.log)
			return
		}
	}

	err = s.CreateEvents(calendar.Events, calUuid, calendar.UserToken, r.Context())
	if err != nil {
		util.WriteInternalServer(w, err, ERR_CREATE_CALENDER, s.log)
		return
	}

	s.log.Infof("created or updated calendar entries for user %s with date %s", calendar.UserToken, time.Unix(calendar.Date, 0).String())

	w.WriteHeader(http.StatusOK)
}
