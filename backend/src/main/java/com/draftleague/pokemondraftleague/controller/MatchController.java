// src/main/java/com/draftleague/pokemondraftleague/controller/MatchController.java
package com.draftleague.pokemondraftleague.controller;

import com.draftleague.pokemondraftleague.model.LeagueStanding;
import com.draftleague.pokemondraftleague.model.Match;
import com.draftleague.pokemondraftleague.service.MatchService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/matches")
public class MatchController {

	private final MatchService matchService;

	public MatchController(MatchService matchService) {
		this.matchService = matchService;
	}

	@PostMapping
	public ResponseEntity<Match> createMatch(@RequestBody Match match) {
		Match savedMatch = matchService.saveMatch(match);
		return new ResponseEntity<>(savedMatch, HttpStatus.CREATED);
	}

	@GetMapping("/{id}")
	public ResponseEntity<Match> getMatchById(@PathVariable Long id) {
		return matchService.getMatchById(id)
				.map(match -> new ResponseEntity<>(match, HttpStatus.OK))
				.orElse(new ResponseEntity<>(HttpStatus.NOT_FOUND));
	}

	@GetMapping("/by-league/{leagueId}")
	public ResponseEntity<List<Match>> getMatchesByLeague(@PathVariable Long leagueId) {
		try {
			List<Match> matches = matchService.getMatchesByLeague(leagueId);
			return new ResponseEntity<>(matches, HttpStatus.OK);
		} catch (IllegalArgumentException e) {
			return new ResponseEntity<>(HttpStatus.NOT_FOUND);
		}
	}

	@PostMapping("/generate-round-robin/{leagueId}")
	public ResponseEntity<List<Match>> generateRoundRobinMatches(@PathVariable Long leagueId) {
		try {
			List<Match> generatedMatches = matchService.generateRoundRobinMatches(leagueId);
			return new ResponseEntity<>(generatedMatches, HttpStatus.CREATED);
		} catch (IllegalArgumentException e) {
			return new ResponseEntity<>(HttpStatus.BAD_REQUEST);
		}
	}

	// Request body for recording match results
	public static class RecordMatchResultRequest {
		public Long winnerId;
		public Integer trainer1Score;
		public Integer trainer2Score;
		public List<String> replayLinks;
	}

	@PutMapping("/{matchId}/record-result")
	public ResponseEntity<Match> recordMatchResult(@PathVariable Long matchId,
			@RequestBody RecordMatchResultRequest request) {
		try {
			return matchService.recordMatchResult(matchId, request.winnerId, request.trainer1Score,
					request.trainer2Score, request.replayLinks)
					.map(match -> new ResponseEntity<>(match, HttpStatus.OK))
					.orElse(new ResponseEntity<>(HttpStatus.NOT_FOUND));
		} catch (IllegalArgumentException e) {
			return new ResponseEntity<>(HttpStatus.BAD_REQUEST);
		}
	}

	@GetMapping("/standings/{leagueId}")
	public ResponseEntity<List<LeagueStanding>> getStandings(@PathVariable Long leagueId) {
		try {
			List<LeagueStanding> standings = matchService.getStandings(leagueId);
			return new ResponseEntity<>(standings, HttpStatus.OK);
		} catch (IllegalArgumentException e) {
			return new ResponseEntity<>(HttpStatus.NOT_FOUND);
		}
	}
}
