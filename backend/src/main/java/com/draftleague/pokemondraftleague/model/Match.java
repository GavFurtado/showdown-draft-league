// src/main/java/com/draftleague/pokemondraftleague/model/Match.java
package com.draftleague.pokemondraftleague.model;

import jakarta.persistence.CollectionTable;
import jakarta.persistence.ElementCollection;
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

import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonBackReference;
import com.fasterxml.jackson.annotation.JsonManagedReference;

@Entity
@Data
@NoArgsConstructor
@AllArgsConstructor
@Table(name = "league_match")
@ToString(exclude = { "league", "trainer1", "trainer2", "winner" })
@EqualsAndHashCode(exclude = { "league", "trainer1", "trainer2", "winner" }) // Exclude related entities from
                                                                             // EqualsAndHashCode
public class Match {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    // A Match belongs to ONE League
    @ManyToOne
    @JoinColumn(name = "league_id", nullable = false) // Foreign key to League
    @JsonBackReference
    private League league;

    // Trainer 1 is a participant in the match
    @ManyToOne
    @JoinColumn(name = "trainer1_id", nullable = false) // Foreign key to Trainer
    @JsonBackReference
    private Trainer trainer1;

    // Trainer 2 is a participant in the match
    @ManyToOne
    @JoinColumn(name = "trainer2_id", nullable = false) // Foreign key to Trainer
    @JsonBackReference
    private Trainer trainer2;

    // The Winner of the match (optional if match not yet played)
    @ManyToOne
    @JoinColumn(name = "winner_id") // Foreign key to Trainer (can be null)
    @JsonBackReference
    private Trainer winner;

    private Integer roundNumber; // e.g., 1, 2, 3... unsure if i'll use this or not
    private String type; // e.g., "REGULAR_SEASON", "PLAYOFF"
    private String status; // e.g., "SCHEDULED", "COMPLETED"

    private Integer trainer1Score;
    private Integer trainer2Score;
    // private java.time.LocalDateTime matchDate; // Requires import
    // java.time.LocalDateTime

    @ElementCollection
    @CollectionTable(name = "match_replay_links", joinColumns = @JoinColumn(name = "match_id"))
    private List<String> showdownReplayLinks = new ArrayList<>();
}
