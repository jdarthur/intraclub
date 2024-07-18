import * as React from "react"
import {isMatchWon} from "./SetScores";
import {MatchupProps} from "./Matchup";
import {
    CheckCircleOutlined,
    ClockCircleFilled,
    ClockCircleOutlined,
    CloseCircleOutlined,
    HistoryOutlined
} from "@ant-design/icons";

type WonLostTopLineProps = {
    Matchups: MatchupProps[]
    Home?: boolean
}

export function WonLostTopLine({Matchups, Home}: WonLostTopLineProps) {

    return Matchups?.map((m) => {
        const us = Home ? m.Result.Us : m.Result.Them
        const them = Home ? m.Result.Them : m.Result.Us


        const won = isMatchWon(us, them)
        const lost = isMatchWon(them, us)

        let v = null
        if (won) {
            v = <CheckCircleOutlined/>
        } else if (lost) {
            v = <CloseCircleOutlined/>
        }
        return <span style={{marginRight: (won || lost) ? "0.25em" : 0}}>
            {v}
        </span>
    })
}