package com.draftleague.pokemondraftleague.repository;

import com.draftleague.pokemondraftleague.model.League;
import com.draftleague.pokemondraftleague.model.LeagueStanding;
import com.draftleague.pokemondraftleague.model.Trainer;
import org.springframework.data.jpa.repository.JpaRepository;
import java.util.List;
import java.util.Optional;

public interface LeagueStandingRepository extends JpaRepository<LeagueStanding, Long> {
    // Find standing for a specific trainer in a specific league
    Optional<LeagueStanding> findByLeagueAndTrainer(League league, Trainer trainer);

    // Find all standings for a given league
    List<LeagueStanding> findByLeague(League league);

    // Find all standings for a given league and order by Descending order
    List<LeagueStanding> findByLeagueOrderByWinsDesc(League league);
}
