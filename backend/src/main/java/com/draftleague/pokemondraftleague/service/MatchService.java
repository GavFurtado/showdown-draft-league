// src/main/java/com/draftleague/pokemondraftleague/service/MatchService.java
package com.draftleague.pokemondraftleague.service;

import com.draftleague.pokemondraftleague.model.League;
import com.draftleague.pokemondraftleague.model.LeagueStanding;
import com.draftleague.pokemondraftleague.model.Match;
import com.draftleague.pokemondraftleague.model.Trainer;
import com.draftleague.pokemondraftleague.repository.MatchRepository;
import com.draftleague.pokemondraftleague.repository.TrainerRepository;
import com.draftleague.pokemondraftleague.repository.LeagueRepository;
import com.draftleague.pokemondraftleague.repository.LeagueStandingRepository;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class MatchService {

    private final MatchRepository matchRepository;
    private final TrainerRepository trainerRepository;
    private final LeagueRepository leagueRepository;
    private final LeagueStandingRepository leagueStandingRepository;

    public MatchService(MatchRepository matchRepository, TrainerRepository trainerRepository,
            LeagueRepository leagueRepository, LeagueStandingRepository leagueStandingRepository) {
        this.matchRepository = matchRepository;
        this.trainerRepository = trainerRepository;
        this.leagueRepository = leagueRepository;
        this.leagueStandingRepository = leagueStandingRepository;
    }

    public Match saveMatch(Match match) {
        return matchRepository.save(match);
    }

    public Optional<Match> getMatchById(Long id) {
        return matchRepository.findById(id);
    }

    public List<Match> getMatchesByLeague(Long leagueId) {
        Optional<League> leagueOptional = leagueRepository.findById(leagueId);
        if (leagueOptional.isEmpty()) {
            throw new IllegalArgumentException("League not found with ID: " + leagueId);
        }
        return matchRepository.findByLeague(leagueOptional.get());
    }

    /**
     * Generates and saves all round-robin matches for a given league.
     * Each trainer plays every other trainer once.
     */
    @Transactional
    public List<Match> generateRoundRobinMatches(Long leagueId) {
        Optional<League> leagueOptional = leagueRepository.findById(leagueId);
        if (leagueOptional.isEmpty()) {
            throw new IllegalArgumentException("League not found with ID: " + leagueId);
        }
        League league = leagueOptional.get();

        List<Trainer> trainers = trainerRepository.findByLeague(league);
        if (trainers.size() < 2) {
            throw new IllegalArgumentException("Not enough trainers in the league to generate matches.");
        }

        List<Match> generatedMatches = new ArrayList<>();
        int matchCount = 0; // Simple counter for round number

        // Basic round robin pairing logic
        for (int i = 0; i < trainers.size(); i++) {
            for (int j = i + 1; j < trainers.size(); j++) {
                Trainer trainer1 = trainers.get(i);
                Trainer trainer2 = trainers.get(j);

                Match match = new Match();
                match.setLeague(league);
                match.setTrainer1(trainer1);
                match.setTrainer2(trainer2);
                match.setRoundNumber(++matchCount);
                match.setType("REGULAR_SEASON");
                match.setStatus("SCHEDULED");
                // showdownReplayLinks, trainer1Score, trainer2Score will be null/empty
                // initially

                generatedMatches.add(match);
            }
        }

        // Initialize LeagueStanding entries for all trainers in this league
        List<LeagueStanding> initialStandings = trainers.stream()
                .map(trainer -> {
                    Optional<LeagueStanding> existingStanding = leagueStandingRepository.findByLeagueAndTrainer(league,
                            trainer);
                    return existingStanding.orElseGet(() -> {
                        LeagueStanding newStanding = new LeagueStanding();
                        newStanding.setLeague(league);
                        newStanding.setTrainer(trainer);
                        newStanding.setWins(0);
                        newStanding.setLosses(0);
                        return newStanding;
                    });
                })
                .collect(Collectors.toList());
        leagueStandingRepository.saveAll(initialStandings); // Save or update existing to ensure all are 0/0

        return matchRepository.saveAll(generatedMatches);
    }

    /**
     * Records the result of a match (sets winner, status, scores, and replay
     * links).
     * 
     * @param matchId       The ID of the match to update.
     * @param winnerId      The ID of the trainer who won.
     * @param trainer1Score The score of Trainer 1 in the series.
     * @param trainer2Score The score of Trainer 2 in the series.
     * @param replayLinks   The list of optional replay links.
     * @return The updated Match object.
     */
    @Transactional
    public Optional<Match> recordMatchResult(Long matchId, Long winnerId, Integer trainer1Score, Integer trainer2Score,
            List<String> replayLinks) {
        Optional<Match> matchOptional = matchRepository.findById(matchId);

        if (matchOptional.isPresent()) {
            Match match = matchOptional.get();
            // Ensure the winnerId is one of the participants
            if (!match.getTrainer1().getId().equals(winnerId) && !match.getTrainer2().getId().equals(winnerId)) {
                throw new IllegalArgumentException("Winner ID does not match any participant in the match.");
            }

            Optional<Trainer> winnerTrainerOptional = trainerRepository.findById(winnerId);
            if (winnerTrainerOptional.isEmpty()) {
                throw new IllegalArgumentException("Winner Trainer not found with ID: " + winnerId);
            }

            // Prevent re-updating standings if the match is already completed
            if ("COMPLETED".equals(match.getStatus())) {
                throw new IllegalStateException(
                        "Match " + matchId + " has already been completed and cannot be re-recorded.");
            }

            Trainer winner = winnerTrainerOptional.get();
            Trainer loser = match.getTrainer1().getId().equals(winnerId) ? match.getTrainer2() : match.getTrainer1();
            League league = match.getLeague();

            // Update winner's standing
            leagueStandingRepository.findByLeagueAndTrainer(league, winner).ifPresent(standing -> {
                standing.setWins(standing.getWins() + 1);
                leagueStandingRepository.save(standing);
            });

            // Update loser's standing
            leagueStandingRepository.findByLeagueAndTrainer(league, loser).ifPresent(standing -> {
                standing.setLosses(standing.getLosses() + 1);
                leagueStandingRepository.save(standing);
            });

            match.setWinner(winnerTrainerOptional.get());
            match.setStatus("COMPLETED");
            match.setTrainer1Score(trainer1Score);
            match.setTrainer2Score(trainer2Score);
            match.setShowdownReplayLinks(replayLinks);
            return Optional.of(matchRepository.save(match));
        }
        return Optional.empty();
    }

    public List<LeagueStanding> getStandings(Long leagueId) { // Change return type
        Optional<League> leagueOptional = leagueRepository.findById(leagueId);
        if (leagueOptional.isEmpty()) {
            throw new IllegalArgumentException("League not found with ID: " + leagueId);
        }
        League league = leagueOptional.get();

        // Directly fetch the standings that are maintained via recordMatchResult
        return leagueStandingRepository.findByLeagueOrderByWinsDesc(league);
    }
}
