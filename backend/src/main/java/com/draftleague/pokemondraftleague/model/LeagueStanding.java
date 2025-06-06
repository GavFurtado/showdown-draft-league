package com.draftleague.pokemondraftleague.model;

import com.fasterxml.jackson.annotation.JsonBackReference;
import com.fasterxml.jackson.annotation.JsonManagedReference;

import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.EqualsAndHashCode;
import lombok.ToString;

@Entity
@Data
@NoArgsConstructor
@AllArgsConstructor
@Table(name = "league_standing", uniqueConstraints = @UniqueConstraint(columnNames = { "league_id", "trainer_id" }) // Ensures
                                                                                                                    // only
                                                                                                                    // one
                                                                                                                    // standing
                                                                                                                    // per
                                                                                                                    // trainer
                                                                                                                    // per
                                                                                                                    // league
)
@ToString(exclude = { "league", "trainer" }) // Exclude related entities to prevent infinite loops
@EqualsAndHashCode(exclude = { "league", "trainer" }) // Exclude related entities from EqualsAndHashCode
public class LeagueStanding {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne
    @JoinColumn(name = "league_id", nullable = false)
    @JsonBackReference
    private League league;

    @ManyToOne
    @JoinColumn(name = "trainer_id", nullable = false)
    @JsonBackReference
    private Trainer trainer;

    private Integer wins = 0; // Initialize to 0 by default
    private Integer losses = 0; // Initialize to 0 by default

    // You could add other metrics here later if needed:
    // private Integer gamesPlayed = 0;
    // private Integer scoreDifferential = 0;
}
