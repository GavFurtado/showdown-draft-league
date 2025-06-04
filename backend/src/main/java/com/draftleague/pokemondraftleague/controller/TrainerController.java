// src/main/java/com/draftleague/pokemondraftleague/controller/TrainerController.java
package com.draftleague.pokemondraftleague.controller;

import com.draftleague.pokemondraftleague.model.Trainer;
import com.draftleague.pokemondraftleague.service.TrainerService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/trainers")
public class TrainerController {

	private final TrainerService trainerService;

	public TrainerController(TrainerService trainerService) {
		this.trainerService = trainerService;
	}

	@GetMapping
	public ResponseEntity<List<Trainer>> getAllTrainers() {
		List<Trainer> trainers = trainerService.getAllTrainers();
		return new ResponseEntity<>(trainers, HttpStatus.OK);
	}

	@GetMapping("/{id}")
	public ResponseEntity<Trainer> getTrainerById(@PathVariable Long id) {
		return trainerService.getTrainerById(id)
				.map(trainer -> new ResponseEntity<>(trainer, HttpStatus.OK))
				.orElse(new ResponseEntity<>(HttpStatus.NOT_FOUND));
	}

	// Request body for registration
	public static class RegisterTrainerRequest {
		public Trainer trainer;
		public Long leagueId;
	}

	@PostMapping("/register")
	public ResponseEntity<Trainer> registerTrainer(@RequestBody RegisterTrainerRequest request) {
		try {
			Trainer registeredTrainer = trainerService.registerTrainer(request.trainer, request.leagueId);
			return new ResponseEntity<>(registeredTrainer, HttpStatus.CREATED);
		} catch (IllegalArgumentException e) {
			return new ResponseEntity<>(HttpStatus.BAD_REQUEST);
		}
	}

	// Request body for login
	public static class LoginRequest {
		public String username;
		public String password;
	}

	@PostMapping("/login")
	public ResponseEntity<Trainer> loginTrainer(@RequestBody LoginRequest request) {
		return trainerService.loginTrainer(request.username, request.password)
				.map(trainer -> new ResponseEntity<>(trainer, HttpStatus.OK))
				.orElse(new ResponseEntity<>(HttpStatus.UNAUTHORIZED)); // 401 Unauthorized
	}

	@DeleteMapping("/{id}")
	public ResponseEntity<Void> deleteTrainer(@PathVariable Long id) {
		if (trainerService.getTrainerById(id).isEmpty()) {
			return new ResponseEntity<>(HttpStatus.NOT_FOUND);
		}
		trainerService.deleteTrainer(id);
		return new ResponseEntity<>(HttpStatus.NO_CONTENT);
	}
}
