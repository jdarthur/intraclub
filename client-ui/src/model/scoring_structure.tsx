export type ScoreCountingType = {
    type: number;
    name: string;
}

export type WinCondition = {
    win_threshold: number;
    must_win_by: number;
    instant_win_threshold: number;
}

export type ScoringStructure = {
    id?: string;
    owner?: string;
    name: string;
    use_instant_win?: boolean;
    win_condition_counting_type: number;
    win_condition: WinCondition;
    is_composite?: boolean;
    secondary_scoring_structures: string[];
}
