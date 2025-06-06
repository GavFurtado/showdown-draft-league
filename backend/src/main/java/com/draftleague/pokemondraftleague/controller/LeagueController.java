// src/main/java/com/draftleague/pokemondraftleague/controller/LeagueController.java
package com.draftleague.pokemondraftleague.controller;

import com.draftleague.pokemondraftleague.model.League;
import com.draftleague.pokemondraftleague.service.LeagueService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/leagues")
public class LeagueController {

    private final LeagueService leagueService;

    public LeagueController(LeagueService leagueService) {
        this.leagueService = leagueService;
    }

    @PostMapping
    public ResponseEntity<League> createLeague(@RequestBody League league) {
        League savedLeague = leagueService.saveLeague(league);
        return new ResponseEntity<>(savedLeague, HttpStatus.CREATED);
    }

    @GetMapping
    public ResponseEntity<List<League>> getAllLeagues() {
        List<League> leagues = leagueService.getAllLeagues();
        return new ResponseEntity<>(leagues, HttpStatus.OK);
    }

    @GetMapping("/{id}")
    public ResponseEntity<League> getLeagueById(@PathVariable Long id) {
        return leagueService.getLeagueById(id)
                .map(league -> new ResponseEntity<>(league, HttpStatus.OK))
                .orElse(new ResponseEntity<>(HttpStatus.NOT_FOUND));
    }

    @PutMapping("/{id}")
    public ResponseEntity<League> updateLeague(@PathVariable Long id, @RequestBody League league) {
        // Ensure the ID in the path matches the ID in the request body for consistency
        if (!league.getId().equals(id)) {
            return new ResponseEntity<>(HttpStatus.BAD_REQUEST);
        }
        // You might want to fetch the existing league first, update its fields, and
        // then save
        // For simplicity, this directly saves, assuming the ID in the payload correctly
        // identifies the entity
        if (leagueService.getLeagueById(id).isEmpty()) {
            return new ResponseEntity<>(HttpStatus.NOT_FOUND);
        }
        League updatedLeague = leagueService.saveLeague(league);
        return new ResponseEntity<>(updatedLeague, HttpStatus.OK);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteLeague(@PathVariable Long id) {
        if (leagueService.getLeagueById(id).isEmpty()) {
            return new ResponseEntity<>(HttpStatus.NOT_FOUND);
        }
        leagueService.deleteLeague(id);
        return new ResponseEntity<>(HttpStatus.NO_CONTENT);
    }
}
