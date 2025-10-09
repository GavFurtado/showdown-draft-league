package services

import (
	"fmt"
	"log"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	u "github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
)

type SchedulerService interface {
	Start() error
	// Stop()
}

type schedulerServiceImpl struct {
	tasks    *u.TaskHeap
	taskChan chan *u.ScheduledTask
	// updateChan chan *u.ScheduledTask // not planned for now
	stopChan   chan struct{}
	leagueRepo repositories.LeagueRepository
	draftRepo  repositories.DraftRepository
}

func NewSchedulerService(
	tasks *u.TaskHeap,
	taskChan chan *u.ScheduledTask,
	stopChan chan struct{},
	leagueRepo repositories.LeagueRepository,
	draftRepo repositories.DraftRepository,
) SchedulerService {
	return &schedulerServiceImpl{
		tasks:      tasks,
		taskChan:   taskChan,
		stopChan:   stopChan,
		leagueRepo: leagueRepo,
		draftRepo:  draftRepo,
	}
}

func (s *schedulerServiceImpl) Start() error {
	// fetch all ongoing drafts
	drafts, err := s.draftRepo.GetAllDraftsByStatus(enums.DraftStatusOngoing)
	if err != nil {
		log.Printf("LOG: (SchedulerService: Start) - error fetching drafts with status %s: %v\n", enums.DraftStatusOngoing, err)
		return err
	}

	// fetch leagues that use the transfer credit system
	leagues, err := s.leagueRepo.GetLeaguesThatAllowTransferCredits()
	if err != nil {
		log.Printf("LOG: (SchedulerService: Start) - error fetching drafts with transfer credit system enabled: %v\n", err)
		return err
	}

	var leaguesInTransferWindow []*models.League
	// Leagues in regular season or those that are bracket only; No transfer credit accrual during playoffs planned
	var leaguesInSeasonOrBracketOnly []*models.League
	for _, league := range leagues {
		if league.Status == enums.LeagueStatusTransferWindow {
			leaguesInTransferWindow = append(leaguesInTransferWindow, &league)
		} else if league.Status == enums.LeagueStatusRegularSeason {
			leaguesInSeasonOrBracketOnly = append(leaguesInSeasonOrBracketOnly, &league)
		} else if league.Format.SeasonType == enums.LeagueSeasonTypeBracketOnly &&
			league.Status == enums.LeagueStatusPlayoffs {
			leaguesInSeasonOrBracketOnly = append(leaguesInSeasonOrBracketOnly, &league)
		}
	}

	// create task objects
	for _, draft := range drafts {
		turnTimeLimit := draft.TurnTimeLimit
		turnStartTime := draft.CurrentTurnStartTime
		turnEndTime := turnStartTime.Add(time.Duration(turnTimeLimit) * time.Minute)

		newTask := &u.ScheduledTask{
			ID:        uuid.New(),
			ExecuteAt: turnEndTime,
			Type:      u.TurnTypeDraftTurnTimeout,
			Payload: u.PayloadDraftTurnTimeout{
				DraftID:  draft.ID,
				LeagueID: draft.LeagueID,
				PlayerID: *draft.CurrentTurnPlayerID,
			},
		}

		s.tasks.Push(newTask)
	}

	for _, league := range leaguesInTransferWindow {
		windowStartTime := league.Format.NextTransferWindowStart
		windowDuration := league.Format.TransferWindowDuration
		windowEndTime := windowStartTime.Add(time.Duration(windowDuration) * time.Minute)

		newTask := &u.ScheduledTask{
			ID:        uuid.New(),
			ExecuteAt: windowEndTime,
			Type:      u.TurnTypeTradingPeriodEnd,
			Payload: u.PayloadTransferPeriodEnd{
				LeagueID: league.ID,
			},
		}
		s.tasks.Push(newTask)
	}

	for _, league := range leaguesInSeasonOrBracketOnly {
		nextWindowStartTime := league.Format.NextTransferWindowStart

		newTask := &u.ScheduledTask{
			ID:        uuid.New(),
			ExecuteAt: *nextWindowStartTime,
			Type:      u.TurnTypeAccrueCredits,
			Payload: u.PayloadTransferCreditAccrual{
				LeagueID: league.ID,
			},
		}
		s.tasks.Push(newTask)
	}

	fmt.Printf("INFO: (SchedulerService: Start) - Running Scheduler\n")
	s.runSchedulerLoop()

	return nil
}

func (s *schedulerServiceImpl) runSchedulerLoop() {

}
