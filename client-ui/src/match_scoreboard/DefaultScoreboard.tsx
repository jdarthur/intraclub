import * as React from "react";
import {Scoreboard} from "./Scoreboard";
import {Team} from "./TeamName";
import {Matchup} from "./Matchup";
import {PairingProps} from "./Pairing";
import {MatchProps} from "./SetScores";

export function DefaultScoreboard() {
    return <div style={{background: "#f5f5f5", height: "100vh"}}>
        <Scoreboard />
    </div>

}