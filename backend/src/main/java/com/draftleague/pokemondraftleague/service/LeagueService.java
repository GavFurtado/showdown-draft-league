// src/main/java/com/draftleague/pokemondraftleague/service/LeagueService.java
package com.draftleague.pokemondraftleague.service;

import com.draftleague.pokemondraftleague.model.League;
import com.draftleague.pokemondraftleague.repository.LeagueRepository;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class LeagueService {

    private final LeagueRepository leagueRepository;

    public LeagueService(LeagueRepository leagueRepository) {
        this.leagueRepository = leagueRepository;
    }

    public League saveLeague(League league) {
        return leagueRepository.save(league);
    }

    public List<League> getAllLeagues() {
        return leagueRepository.findAll();
    }

    public Optional<League> getLeagueById(Long id) {
        return leagueRepository.findById(id);
    }

    public void deleteLeague(Long id) {
        leagueRepository.deleteById(id);
    }
}
