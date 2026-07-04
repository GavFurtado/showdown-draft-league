package services

import (
	"container/heap"
	"fmt"
	"log/slog"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
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
	logger          *slog.Logger
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
	logger *slog.Logger,
	tasks *u.TaskHeap,
	leagueRepo repositories.LeagueRepository,
	draftRepo repositories.DraftRepository,
) SchedulerService {
	return &schedulerServiceImpl{
		logger:         utils.LoggerWithService(logger, "SchedulerService"),
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
	// Fetch all ongoing drafts
	drafts, err := s.draftRepo.GetAllDraftsByStatus(enums.DraftStatusOngoing)
	if err != nil {
		s.logger.Error("Start - error fetching drafts with status", "status", enums.DraftStatusOngoing, "error", err)
		return err
	}

	// Fetch leagues that use the transfer credit system
	leagues, err := s.leagueRepo.GetLeaguesThatAllowTransfers()
	if err != nil {
		s.logger.Error("Start - error fetching leagues with transfers enabled", "error", err)
		return err
	}

	// Schedule LeagueWeeklyTick for ongoing regular season leagues and leagues currently in transfer window
	// If the server restarts during a transfer window, we still need to make sure the next tick is scheduled.
	ongoingLeagues, err := s.leagueRepo.GetLeaguesByStatuses([]enums.LeagueStatus{enums.LeagueStatusRegularSeason, enums.LeagueStatusTransferWindow})
	if err != nil {
		s.logger.Error("Start - error fetching ongoing regular season/transfer window leagues", "error", err)
		return err
	}
	for _, league := range ongoingLeagues {
		if league.NextWeeklyTick != nil {
			// If a tick is in the past, execute it immediately. Otherwise, schedule it for its designated time.
			executeAt := *league.NextWeeklyTick
			if executeAt.Before(time.Now()) {
				s.logger.Warn("Start - weekly tick for league is overdue, scheduling for immediate execution", "league_id", league.ID)
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
			heap.Push(s.tasks, newTask)
			s.taskMap[newTask.ID] = newTask
			s.logger.Info("Start - restored weekly tick for league", "league_id", league.ID, "execute_at", executeAt.String())
		}
	}

	var leaguesInTransferWindow []*models.League
	// Leagues in regular season or those that are bracket only; No transfer credit accrual during playoffs planned
	var leaguesInSeasonOrBracketOnly []*models.League
	for _, league := range leagues {
		if league.Format == nil {
			s.logger.Warn("Start - league has nil Format, skipping", "league_id", league.ID)
			continue
		}
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
				PlayerID: *draft.CurrentTurnMemberID,
			},
		}
		heap.Push(s.tasks, newTask)
		s.taskMap[newTask.ID] = newTask
	}

	for _, league := range leaguesInTransferWindow {
		if league.Format.NextTransferWindowStart == nil {
			s.logger.Warn("Start - league is in TransferWindow but NextTransferWindowStart is nil, skipping", "league_id", league.ID)
			continue
		}
		windowStartTime := league.Format.NextTransferWindowStart
		windowDuration := league.Format.TransferWindowDuration
		windowEndTime := windowStartTime.Add(time.Duration(windowDuration) * time.Minute)

		newTask := &u.ScheduledTask{
			ID:        fmt.Sprintf("%d_%s", u.TaskTypeTransferPeriodEnd, league.ID),
			ExecuteAt: windowEndTime,
			Type:      u.TaskTypeTransferPeriodEnd,
			Payload: u.PayloadTransferPeriodEnd{
				LeagueID: league.ID,
			},
		}
		heap.Push(s.tasks, newTask)
		s.taskMap[newTask.ID] = newTask
	}

	for _, league := range leaguesInSeasonOrBracketOnly {
		if league.Format.NextTransferWindowStart == nil {
			s.logger.Warn("Start - league is in Season/Bracket but NextTransferWindowStart is nil, skipping", "league_id", league.ID)
			continue
		}
		nextWindowStartTime := league.Format.NextTransferWindowStart

		newTask := &u.ScheduledTask{
			ID:        fmt.Sprintf("%d_%s", u.TaskTypeTransferPeriodStart, league.ID),
			ExecuteAt: *nextWindowStartTime,
			Type:      u.TaskTypeTransferPeriodStart,
			Payload: u.PayloadTransferPeriodStart{
				LeagueID: league.ID,
			},
		}
		heap.Push(s.tasks, newTask)
		s.taskMap[newTask.ID] = newTask
	}

	s.logger.Info("Start - running scheduler")
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
	s.logger.Info("RegisterTask - task registered", "task_id", task.ID, "type", task.Type, "execute_at", task.ExecuteAt)
}

// runSchedulerLoop is the main loop of the scheduler that processes tasks.
func (s *schedulerServiceImpl) runSchedulerLoop() {
	var timer *time.Timer
	for {
		now := time.Now()
		upcomingTask, exists := s.tasks.Peek()

		if exists { // if there was a task
			if upcomingTask.ExecuteAt.Before(now) {
				// task is overdue; execute now
				s.logger.Info("a task is overdue, executing now")
				timer = time.NewTimer(0) // fire new timer immediately to execute task
			} else {
				// the task is not due yet; wait till due
				waitDuration := upcomingTask.ExecuteAt.Sub(now)
				timer = time.NewTimer(waitDuration)
				s.logger.Info("task(s) are scheduled but not due", "wait_duration", waitDuration)
				s.tasks.Print()
			}
		} else {
			// no tasks on the priority queue, wait for a task
			s.logger.Info("no tasks on the queue, waiting")
			timer = time.NewTimer(time.Hour * 24 * 365 * 10) // long ahh time
		}

		select {
		case newTask := <-s.taskChan:
			// a new task has been submitted by another service
			s.logger.Info("scheduler received a new task", "task_id", newTask.ID, "type", newTask.Type, "execute_at", newTask.ExecuteAt)
			heap.Push(s.tasks, newTask)
		case <-s.rescheduleChan:
			s.logger.Info("reschedule signal received, re-evaluating next task")
			// nothing else needs to be done here. timer will be rescheduled in the following iteration
			continue
		case <-timer.C:
			// timer fired; execute the scheduled task
			task := heap.Pop(s.tasks).(*u.ScheduledTask)
			s.logger.Info("scheduler executing task", "task_id", task.ID, "type", task.Type, "execute_at", task.ExecuteAt)
			// Execute the task using the injected DraftTaskExecutor
			s.executeTask(task)
			delete(s.taskMap, task.ID)
		case <-s.stopChan:
			// currently nothing sends a signal to this channel
			// Stop() call was made
			s.logger.Info("scheduler received stop signal, shutting down")
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
		s.logger.Warn("DeregisterTask - attempted to deregister non-existent task", "task_id", taskID)
		return
	}

	// Remove from the heap
	heap.Remove(s.tasks, task.Index)

	// Remove from the map
	delete(s.taskMap, taskID)
	s.logger.Info("DeregisterTask - task deregistered", "task_id", taskID)

	s.rescheduleChan <- struct{}{}
}

// executeTask checks the type of the task to execute then makes the appropriate execute call for the task
func (s *schedulerServiceImpl) executeTask(task *u.ScheduledTask) {
	switch task.Type {
	case u.TaskTypeDraftTurnTimeout:
		if payload, ok := task.Payload.(u.PayloadDraftTurnTimeout); ok {
			s.logger.Info("executeTask - draft turn timeout", "league_id", payload.LeagueID, "player_id", payload.PlayerID)
			if s.draftService == nil {
				s.logger.Error("executeTask - DraftService is not set, cannot auto-skip turn", "league_id", payload.LeagueID, "player_id", payload.PlayerID)
				return
			}

			if err := s.draftService.AutoSkipTurn(payload.PlayerID, payload.LeagueID); err != nil {
				s.logger.Error("executeTask - error occurred in AutoSkipTurn", "error", err)
				return
			}
		} else {
			s.logger.Error("executeTask - invalid payload type for DraftTurnTimeout task", "task_id", task.ID)
		}

	case u.TaskTypeTransferPeriodEnd:
		if payload, ok := task.Payload.(u.PayloadTransferPeriodEnd); ok {
			s.logger.Info("executeTask - transfer period end", "league_id", payload.LeagueID)
			if s.transferService == nil {
				s.logger.Error("executeTask - TransferService is not set, cannot end transfer period", "league_id", payload.LeagueID)
				return
			}

			if err := s.transferService.EndTransferPeriod(payload.LeagueID); err != nil {
				s.logger.Error("executeTask - error occurred in EndTransferPeriod", "error", err)
				return
			}
		} else {
			s.logger.Error("executeTask - invalid payload type for TransferPeriodEnd task", "task_id", task.ID)
		}

	case u.TaskTypeTransferPeriodStart:
		if payload, ok := task.Payload.(u.PayloadTransferPeriodStart); ok {
			s.logger.Info("executeTask - transfer window start", "league_id", payload.LeagueID)
			if s.transferService == nil {
				s.logger.Error("executeTask - TransferService is not set, cannot start transfer period", "league_id", payload.LeagueID)
				return
			}
			if err := s.transferService.StartTransferPeriod(payload.LeagueID); err != nil {
				s.logger.Error("executeTask - error occurred in StartTransferPeriod", "error", err)
				return
			}
		} else {
			s.logger.Error("executeTask - invalid payload type for StartTransferPeriod task", "task_id", task.ID)
		}
	case u.TaskTypeLeagueWeeklyTick:
		if payload, ok := task.Payload.(u.PayloadLeagueWeeklyTick); ok {
			s.logger.Info("executeTask - league weekly tick", "league_id", payload.LeagueID)
			if s.leagueService == nil {
				s.logger.Error("executeTask - LeagueService is not set, cannot process weekly tick", "league_id", payload.LeagueID)
				return
			}
			if err := s.leagueService.ProcessWeeklyTick(payload.LeagueID); err != nil {
				s.logger.Error("executeTask - error occurred in ProcessWeeklyTick", "error", err)
				return
			}
		} else {
			s.logger.Error("executeTask - invalid payload type for LeagueWeeklyTick task", "task_id", task.ID)
		}
	default:
		s.logger.Error("executeTask - unknown task type", "task_type", task.Type, "task_id", task.ID)
	}
}

// Stop gracefully shuts down the scheduler's background goroutine.
func (s *schedulerServiceImpl) Stop() {
	// Just sends struct{} to the stopChan
	// which will shut down the go routine
	s.stopChan <- struct{}{}
}
