package services

import (
	"container/heap"
	"fmt"
	"log"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	u "github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
)

// SchedulerService defines the interface for managing scheduled tasks.
type SchedulerService interface {
	Start() error
	RegisterTask(task *u.ScheduledTask)
	DeregisterTask(taskID string)
	Stop()
	SetDraftService(draftService DraftService)
	SetTransferService(transferService TransferService)
	SetLeagueService(leagueService LeagueService)
}

type schedulerServiceImpl struct {
	tasks           *u.TaskHeap
	taskMap         map[string]*u.ScheduledTask
	taskChan        chan *u.ScheduledTask
	rescheduleChan  chan struct{}
	stopChan        chan struct{}
	leagueRepo      repositories.LeagueRepository
	draftRepo       repositories.DraftRepository
	draftService    DraftService
	transferService TransferService
	leagueService   LeagueService
}

func NewSchedulerService(
	tasks *u.TaskHeap,
	leagueRepo repositories.LeagueRepository,
	draftRepo repositories.DraftRepository,
) SchedulerService {
	return &schedulerServiceImpl{
		tasks:          tasks,
		taskMap:        make(map[string]*u.ScheduledTask),
		taskChan:       make(chan *u.ScheduledTask, 5),
		rescheduleChan: make(chan struct{}, 1),
		stopChan:       make(chan struct{}),
		leagueRepo:     leagueRepo,
		draftRepo:      draftRepo,
	}
}

// SetDraftService injects the dependency needed for the scheduler to execute draft-related tasks.
// This is set during application startup to break the circular dependency with DraftService.
func (s *schedulerServiceImpl) SetDraftService(draftService DraftService) {
	s.draftService = draftService
}

// SetTransferService injects the dependency needed for the scheduler to execute transfer-related tasks.
// This is set during application startup to break the circular dependency with TransferServer.
func (s *schedulerServiceImpl) SetTransferService(transferService TransferService) {
	s.transferService = transferService
}

// SetLeagueService injects the dependency needed for the scheduler to execute league-related tasks.
// This is set during application startup to break the circular dependency with LeagueService.
func (s *schedulerServiceImpl) SetLeagueService(leagueService LeagueService) {
	s.leagueService = leagueService
}

// Start initializes the scheduler on application boot. It fetches all ongoing drafts
// and active league phases from the database, reconstructs the necessary tasks
// (e.g.,turn timeouts), and launches the main scheduling loop in a background goroutine.
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
		log.Printf("LOG: (SchedulerService: Start) - error fetching leagues with transfer credit system enabled: %v\n", err)
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
			ID:        fmt.Sprintf("%d_%s", u.TaskTypeDraftTurnTimeout, draft.ID),
			ExecuteAt: turnEndTime,
			Type:      u.TaskTypeDraftTurnTimeout,
			Payload: u.PayloadDraftTurnTimeout{
				DraftID:  draft.ID,
				LeagueID: draft.LeagueID,
				PlayerID: *draft.CurrentTurnPlayerID,
			},
		}
		s.tasks.Push(newTask)
		s.taskMap[newTask.ID] = newTask
	}

	for _, league := range leaguesInTransferWindow {
		windowStartTime := league.Format.NextTransferWindowStart
		windowDuration := league.Format.TransferWindowDuration
		windowEndTime := windowStartTime.Add(time.Duration(windowDuration) * time.Minute)

		newTask := &u.ScheduledTask{
			ID:        fmt.Sprintf("%d_%s", u.TaskTypeTradingPeriodEnd, league.ID),
			ExecuteAt: windowEndTime,
			Type:      u.TaskTypeTradingPeriodEnd,
			Payload: u.PayloadTransferPeriodEnd{
				LeagueID: league.ID,
			},
		}
		s.tasks.Push(newTask)
		s.taskMap[newTask.ID] = newTask
	}

	for _, league := range leaguesInSeasonOrBracketOnly {
		nextWindowStartTime := league.Format.NextTransferWindowStart

		newTask := &u.ScheduledTask{
			ID:        fmt.Sprintf("%d_%s", u.TaskTypeTradingPeriodStart, league.ID),
			ExecuteAt: *nextWindowStartTime,
			Type:      u.TaskTypeTradingPeriodStart,
			Payload: u.PayloadTransferPeriodStart{
				LeagueID: league.ID,
			},
		}
		s.tasks.Push(newTask)
		s.taskMap[newTask.ID] = newTask
	}

	// Schedule LeagueWeeklyTick for ongoing regular season leagues
	ongoingRegularSeasonLeagues, err := s.leagueRepo.GetAllLeaguesByStatus(enums.LeagueStatusRegularSeason)
	if err != nil {
		log.Printf("LOG: (SchedulerService: Start) - error fetching ongoing regular season leagues: %v\n", err)
		return err
	}

	for _, league := range ongoingRegularSeasonLeagues {
		if league.NextWeeklyTick != nil {
			// If a tick is in the past, execute it immediately. Otherwise, schedule it for its designated time.
			executeAt := *league.NextWeeklyTick
			if executeAt.Before(time.Now()) {
				log.Printf("LOG: (SchedulerService: Start) - Weekly tick for league %s is overdue. Scheduling for immediate execution.\n", league.ID)
				executeAt = time.Now()
			}

			newTask := &u.ScheduledTask{
				ID:        fmt.Sprintf("%d_%s", u.TaskTypeLeagueWeeklyTick, league.ID),
				ExecuteAt: executeAt,
				Type:      u.TaskTypeLeagueWeeklyTick,
				Payload: u.PayloadLeagueWeeklyTick{
					LeagueID: league.ID,
				},
			}
			s.tasks.Push(newTask)
			s.taskMap[newTask.ID] = newTask
			log.Printf("LOG: (SchedulerService: Start) - Restored weekly tick for league %s, scheduled for %s.\n", league.ID, executeAt.String())
		}
	}

	log.Printf("LOG: (SchedulerService: Start) - Running Scheduler\n")
	go s.runSchedulerLoop()

	return nil
}

// RegisterTask adds a new task to the scheduler. It is called by other services
// to schedule a future action, such as the timeout for a draft turn.
func (s *schedulerServiceImpl) RegisterTask(task *u.ScheduledTask) {
	// add to the map for quick lookup and deregistration
	s.taskMap[task.ID] = task
	// send to the channel for the scheduler loop to pick up
	s.taskChan <- task
	log.Printf("LOG: (SchedulerService: RegisterTask) - Task registered: %s (Type: %s, ExecuteAt: %s)\n", task.ID, task.Type, task.ExecuteAt)
}

// runSchedulerLoop is the main loop of the scheduler that processes tasks.
func (s *schedulerServiceImpl) runSchedulerLoop() {
	var timer *time.Timer
	for {
		now := time.Now()
		nextTask, exists := s.tasks.Peek()

		if exists { // if there was a task
			if nextTask.ExecuteAt.Before(now) {
				// task is overdue; execute now
				log.Printf("LOG: (SchedulerService: runSchedulerLoop) - A task is overdue. Executing now...\n")
				timer = time.NewTimer(0) // fire new timer immediately to execute task
			} else {
				// the task is not due yet; wait till due
				waitDuration := nextTask.ExecuteAt.Sub(now)
				timer = time.NewTimer(waitDuration)
				log.Printf("LOG: (SchedulerService: runSchedulerLoop) - Task(s) are scheduled but not due. Earliest due task in: %s\n", waitDuration)
			}
		} else {
			// no tasks on the priority queue, wait for a task
			log.Printf("LOG: (SchedulerService: runSchedulerLoop) - no tasks on the queue, waiting...\n")
			timer = time.NewTimer(time.Hour * 24 * 365 * 10) // long ahh time
		}

		select {
		case newTask := <-s.taskChan:
			// a new task has been submitted by another service
			log.Printf("LOG: Scheduler recieved a new task: %s (Type: %s, ExecuteAt: %s)\n", newTask.ID, newTask.Type, newTask.ExecuteAt)
			s.tasks.Push(newTask)
		case <-s.rescheduleChan:
			log.Println("LOG: (SchedulerService: runSchedulerLoop) - Reschedule signal received. Re-evaluating next task.")
			// nothing else needs to be done here. timer will be rescheduled in the following iteration
			continue
		case <-timer.C:
			// timer fired; execute the scheduled task
			task := s.tasks.Pop().(*u.ScheduledTask)
			log.Printf("LOG: Scheduler executing task: %s (Type: %s, ExecuteAt: %s)\n", task.ID, task.Type, task.ExecuteAt)
			// Execute the task using the injected DraftTaskExecutor
			s.executeTask(task)
			delete(s.taskMap, task.ID)
		case <-s.stopChan:
			// currently nothing sends a signal to this channel
			// Stop() call was made
			log.Printf("LOG: Scheduler received stop signal. Shutting Down. Scheduler can be restarted by restarting the server.\n")
			if timer != nil {
				timer.Stop()
			}
			return // stop goroutine
		}
	}
}

// DeregisterTask removes a task from the scheduler. This is called when a task
// is completed ahead of schedule, for example, when a player makes a draft pick
// before their turn timer expires.
func (s *schedulerServiceImpl) DeregisterTask(taskID string) {
	task, exists := s.taskMap[taskID]
	if !exists {
		log.Printf("WARN: (SchedulerService: DeregisterTask) - Attempted to deregister non-existent task: %s\n", taskID)
		return
	}

	// Remove from the heap
	heap.Remove(s.tasks, task.Index)

	// Remove from the map
	delete(s.taskMap, taskID)
	log.Printf("LOG: (SchedulerService: DeregisterTask) - Task deregistered: %s\n", taskID)

	s.rescheduleChan <- struct{}{}
}

// executeTask checks the type of the task to execute then makes the appropriate execute call for the task
func (s *schedulerServiceImpl) executeTask(task *u.ScheduledTask) {
	switch task.Type {
	case u.TaskTypeDraftTurnTimeout:
		if payload, ok := task.Payload.(u.PayloadDraftTurnTimeout); ok {
			log.Printf("LOG: (SchedulerService: executeTask) - Draft turn timeout for LeagueID: %s, PlayerID: %s\n", payload.LeagueID, payload.PlayerID)
			if s.draftService == nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - DraftService is not set. Cannot auto-skip turn for LeagueID: %s, PlayerID: %s\n", payload.LeagueID, payload.PlayerID)
				return
			}

			if err := s.draftService.AutoSkipTurn(payload.PlayerID, payload.LeagueID); err != nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - error occured in AutoSkipTurn: %v\n", err)
				return
			}
		} else {
			log.Printf("ERROR: (SchedulerService: executeTask) - Invalid payload type for DraftTurnTimeout task ID %s\n", task.ID)
		}

	case u.TaskTypeTradingPeriodEnd:
		if payload, ok := task.Payload.(u.PayloadTransferPeriodEnd); ok {
			log.Printf("LOG: (SchedulerService: executeTask) - Transfer period end for LeagueID: %s\n", payload.LeagueID)
			if s.transferService == nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - TransferService is not set. Cannot end transfer period for LeagueID: %s\n", payload.LeagueID)
				return
			}

			if err := s.transferService.EndTransferPeriod(payload.LeagueID); err != nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - error occured in EndTransferPeriod: %v\n", err)
				return
			}
		} else {
			log.Printf("ERROR: (SchedulerService: executeTask) - Invalid payload type for TransferPeriodEnd task ID %s\n", task.ID)
		}

	case u.TaskTypeTradingPeriodStart:
		if payload, ok := task.Payload.(u.PayloadTransferPeriodStart); ok {
			log.Printf("LOG: (SchedulerService: executeTask) - Accrue credits for LeagueID: %s\n", payload.LeagueID)
			if s.transferService == nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - TransferService is not set. Cannot start transfer period for LeagueID: %s\n", payload.LeagueID)
				return
			}
			if err := s.transferService.StartTransferPeriod(payload.LeagueID); err != nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - error occured in StartTransferPeriod: %v\n", err)
				return
			}
		} else {
			log.Printf("ERROR: (SchedulerService: executeTask) - Invalid payload type for AccrueCredits task ID %s. Expected PayloadTransferCreditAccrual.\n", task.ID)
		}
	case u.TaskTypeLeagueWeeklyTick:
		if payload, ok := task.Payload.(u.PayloadLeagueWeeklyTick); ok {
			log.Printf("LOG: (SchedulerService: executeTask) - League weekly tick for LeagueID: %s\n", payload.LeagueID)
			if s.leagueService == nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - LeagueService is not set. Cannot process weekly tick for LeagueID: %s\n", payload.LeagueID)
				return
			}
			if err := s.leagueService.ProcessWeeklyTick(payload.LeagueID); err != nil {
				log.Printf("ERROR: (SchedulerService: executeTask) - error occurred in ProcessWeeklyTick: %v\n", err)
				return
			}
		} else {
			log.Printf("ERROR: (SchedulerService: executeTask) - Invalid payload type for LeagueWeeklyTick task ID %s.\n", task.ID)
		}
	default:
		log.Printf("ERROR: (SchedulerService: executeTask) - Unknown task type: %d for task ID %s\n", task.Type, task.ID)
	}
}

// Stop gracefully shuts down the scheduler's background goroutine.
func (s *schedulerServiceImpl) Stop() {
	// Just sends struct{} to the stopChan
	// which will shut down the go routine
	s.stopChan <- struct{}{}
}
