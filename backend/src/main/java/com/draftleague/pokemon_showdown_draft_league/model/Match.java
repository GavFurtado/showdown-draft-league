// src/main/java/com/draftleague/pokemondraftleague/model/Match.java
package com.draftleague.pokemondraftleague.model;

import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn; // For foreign key
import jakarta.persistence.ManyToOne; // For many-to-one relationship
import jakarta.persistence.Table; // To rename the table if 'match' is reserved

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode; // For Lombok
import lombok.NoArgsConstructor;
import lombok.ToString; // For Lombok

@Entity
@Data
@NoArgsConstructor
@AllArgsConstructor
@Table(name = "league_match") // 'match' can be a reserved keyword in some databases, so using 'league_match'
@ToString(exclude = {"league", "trainer1", "trainer2", "winner"}) // Exclude related entities to prevent infinite loops
@EqualsAndHashCode(exclude = {"league", "trainer1", "trainer2", "winner"}) // Exclude related entities from EqualsAndHashCode
public class Match {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    // A Match belongs to ONE League
    @ManyToOne
    @JoinColumn(name = "league_id", nullable = false) // Foreign key to League
    private League league;

    // Trainer 1 is a participant in the match
    @ManyToOne
    @JoinColumn(name = "trainer1_id", nullable = false) // Foreign key to Trainer
    private Trainer trainer1;

    // Trainer 2 is a participant in the match
    @ManyToOne
    @JoinColumn(name = "trainer2_id", nullable = false) // Foreign key to Trainer
    private Trainer trainer2;

    // The Winner of the match (optional if match not yet played)
    @ManyToOne
    @JoinColumn(name = "winner_id") // Foreign key to Trainer (can be null)
    private Trainer winner;

    private Integer roundNumber; // e.g., 1, 2, 3... unsure if i'll use this or not
    private String type; // e.g., "REGULAR_SEASON", "PLAYOFF"
    private String status; // e.g., "SCHEDULED", "COMPLETED"

    // Optional: add fields for scores, date, time if you need them
    // private Integer trainer1Score;
    // private Integer trainer2Score;
    // private java.time.LocalDateTime matchDate; // Requires import java.time.LocalDateTime

    private String showdownReplayLink; // e.g., "https://replay.pokemonshowdown.com/gen9ou-1234567890"
}
