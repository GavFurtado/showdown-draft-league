// src/main/java/com/draftleague/pokemondraftleague/repository/LeagueRepository.java
package com.draftleague.pokemondraftleague.repository;

import com.draftleague.pokemondraftleague.model.League;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface LeagueRepository extends JpaRepository<League, Long> {
    // currently unused because we're working with just the one league
}
