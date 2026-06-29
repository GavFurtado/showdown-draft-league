package services_test

import (
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	mock_repositories "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	mock_services "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type draftServiceMocks struct {
	draftRepo        *mock_repositories.MockDraftRepository
	leagueRepo       *mock_repositories.MockLeagueRepository
	leagueMemberRepo *mock_repositories.MockLeagueMemberRepository
	schedulerService *mock_services.MockSchedulerService

	poolEntryRepo *mock_repositories.MockPoolEntryRepository
	draftPickRepo *mock_repositories.MockDraftPickRepository
	claimRepo     *mock_repositories.MockClaimRepository
}

func setupDraftServiceTest() (services.DraftService, draftServiceMocks) {
	mocks := draftServiceMocks{
		draftRepo:        new(mock_repositories.MockDraftRepository),
		leagueRepo:       new(mock_repositories.MockLeagueRepository),
		leagueMemberRepo: new(mock_repositories.MockLeagueMemberRepository),
		schedulerService: new(mock_services.MockSchedulerService),
		poolEntryRepo:    new(mock_repositories.MockPoolEntryRepository),
		draftPickRepo:    new(mock_repositories.MockDraftPickRepository),
		claimRepo:        new(mock_repositories.MockClaimRepository),
	}

	service := services.NewDraftService(
		mocks.leagueRepo,
		mocks.draftRepo,
		mocks.leagueMemberRepo,
		nil,
	)
	service.SetSchedulerService(mocks.schedulerService)
	service.SetNewRepositories(
		mocks.draftPickRepo,
		mocks.claimRepo,
		mocks.poolEntryRepo,
	)

	return service, mocks
}

func TestDraftService_MakePick(t *testing.T) {
	service, mocks := setupDraftServiceTest()

	// --- Test Data ---
	leagueID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	poolEntryID := uuid.New()
	pokemonSpeciesID := int64(1)

	t.Run("Success - Make a valid pick", func(t *testing.T) {
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              &types.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localMember := &models.LeagueMember{ID: memberID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnMemberID:         &memberID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &requests.DraftMakePickRequestDTO{
			RequestedPickCount: 1,
			RequestedPicks: []requests.RequestedPickDTO{
				{PoolEntryID: poolEntryID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}
		localPoolEntry := &models.PoolEntry{
			ID:               poolEntryID,
			LeagueID:         leagueID,
			PokemonSpeciesID: pokemonSpeciesID,
			Cost:             func(i int) *int { return &i }(50),
			IsAvailable:      true,
		}

		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.leagueMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(localMember, nil).Once()
		mocks.poolEntryRepo.On("GetByIDs", leagueID, []uuid.UUID{poolEntryID}).Return([]models.PoolEntry{*localPoolEntry}, nil).Once()
		mocks.leagueMemberRepo.On("GetCountByLeague", leagueID).Return(int64(1), nil).Once()
		mocks.draftPickRepo.On("CreateBatch", mock.Anything).Return(nil).Once()
		mocks.poolEntryRepo.On("MarkUnavailable", mock.Anything, poolEntryID).Return(nil).Once()
		mocks.leagueMemberRepo.On("Update", mock.AnythingOfType("*models.LeagueMember")).Return(&models.LeagueMember{}, nil).Once()
		mocks.claimRepo.On("Create", mock.AnythingOfType("*models.Claim")).Return(&models.Claim{}, nil).Once()
		mocks.claimRepo.On("GetActiveCountByLeague", leagueID).Return(int64(1), nil).Once()
		mocks.leagueMemberRepo.On("GetByLeague", leagueID).Return([]models.LeagueMember{*localMember}, nil).Once()
		mocks.draftRepo.On("UpdateDraft", mock.AnythingOfType("*models.Draft")).Return(localDraft, nil).Once()
		mocks.schedulerService.On("DeregisterTask", mock.AnythingOfType("string")).Return().Once()
		mocks.schedulerService.On("RegisterTask", mock.AnythingOfType("*utils.ScheduledTask")).Return().Once()

		err := service.MakePick(localCurrentUser, leagueID, localInput)

		assert.NoError(t, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.leagueMemberRepo.AssertExpectations(t)
		mocks.poolEntryRepo.AssertExpectations(t)
		mocks.draftPickRepo.AssertExpectations(t)
		mocks.claimRepo.AssertExpectations(t)
		mocks.schedulerService.AssertExpectations(t)
	})

	t.Run("Failure - Not player's turn", func(t *testing.T) {
		otherMemberID := uuid.New()
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              &types.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localMember := &models.LeagueMember{ID: memberID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:            leagueID,
			Status:              enums.DraftStatusOngoing,
			CurrentPickOnClock:  1,
			CurrentTurnMemberID: &otherMemberID,
		}
		localInput := &requests.DraftMakePickRequestDTO{
			RequestedPickCount: 1,
			RequestedPicks: []requests.RequestedPickDTO{
				{PoolEntryID: poolEntryID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}

		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.leagueMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(localMember, nil).Once()

		err := service.MakePick(localCurrentUser, leagueID, localInput)
		assert.Error(t, err)
		assert.Equal(t, types.ErrUnauthorized, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.leagueMemberRepo.AssertExpectations(t)
	})

	t.Run("Failure - League not in drafting status", func(t *testing.T) {
		localLeague := &models.League{
			ID:     leagueID,
			Status: enums.LeagueStatusSetup,
		}
		localMember := &models.LeagueMember{ID: memberID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnMemberID:         &memberID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &requests.DraftMakePickRequestDTO{
			RequestedPickCount: 1,
			RequestedPicks: []requests.RequestedPickDTO{
				{PoolEntryID: poolEntryID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}

		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.leagueMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(localMember, nil).Once()

		err := service.MakePick(localCurrentUser, leagueID, localInput)

		assert.Error(t, err)
		assert.Equal(t, types.ErrInvalidState, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.leagueMemberRepo.AssertExpectations(t)
	})

	t.Run("Failure - Pokemon not available", func(t *testing.T) {
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              &types.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localMember := &models.LeagueMember{ID: memberID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnMemberID:         &memberID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &requests.DraftMakePickRequestDTO{
			RequestedPickCount: 1,
			RequestedPicks: []requests.RequestedPickDTO{
				{PoolEntryID: poolEntryID, DraftPickNumber: 1},
			},
		}
		unavailablePoolEntry := &models.PoolEntry{
ID:               poolEntryID,
			IsAvailable: false,
		}

		localCurrentUser := &models.User{ID: userID}

		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.leagueMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(localMember, nil).Once()
		mocks.poolEntryRepo.On("GetByIDs", leagueID, []uuid.UUID{poolEntryID}).Return([]models.PoolEntry{*unavailablePoolEntry}, nil).Once()

		err := service.MakePick(localCurrentUser, leagueID, localInput)

		assert.Error(t, err)
		assert.Equal(t, types.ErrConflict, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.leagueMemberRepo.AssertExpectations(t)
		mocks.poolEntryRepo.AssertExpectations(t)
	})

	t.Run("Failure - Insufficient draft points", func(t *testing.T) {
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              &types.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localMember := &models.LeagueMember{ID: memberID, UserID: userID, LeagueID: leagueID, DraftPoints: 20}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnMemberID:         &memberID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &requests.DraftMakePickRequestDTO{
			RequestedPickCount: 1,
			RequestedPicks: []requests.RequestedPickDTO{
				{PoolEntryID: poolEntryID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}
		localPoolEntry := &models.PoolEntry{
			ID:               poolEntryID,
			LeagueID:         leagueID,
			PokemonSpeciesID: pokemonSpeciesID,
			Cost:             func(i int) *int { return &i }(50),
			IsAvailable:      true,
		}

		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.leagueMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(localMember, nil).Once()
		mocks.poolEntryRepo.On("GetByIDs", leagueID, []uuid.UUID{poolEntryID}).Return([]models.PoolEntry{*localPoolEntry}, nil).Once()
		mocks.leagueMemberRepo.On("GetCountByLeague", leagueID).Return(int64(8), nil).Once()

		err := service.MakePick(localCurrentUser, leagueID, localInput)

		assert.Error(t, err)
		assert.Equal(t, types.ErrInsufficientDraftPoints, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.leagueMemberRepo.AssertExpectations(t)
		mocks.poolEntryRepo.AssertExpectations(t)
	})
}
