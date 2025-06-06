// src/main/java/com/draftleague/pokemondraftleague/repository/MatchRepository.java
package com.draftleague.pokemondraftleague.repository;

import com.draftleague.pokemondraftleague.model.Match;
import com.draftleague.pokemondraftleague.model.League; // Import League for filtering
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface MatchRepository extends JpaRepository<Match, Long> {
    // Find all matches for a specific league
    List<Match> findByLeague(League league);

    // Find matches by league and match type (e.g., "REGULAR_SEASON", "PLAYOFF")
    List<Match> findByLeagueAndType(League league, String type);

    // Find matches by league and status (e.g., "SCHEDULED", "COMPLETED")
    List<Match> findByLeagueAndStatus(League league, String status);

    // Custom query method: Find completed regular season matches for standings
    // calculation
    List<Match> findByLeagueAndTypeAndStatus(League league, String type, String status);
}
