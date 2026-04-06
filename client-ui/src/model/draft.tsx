
type TeamCaptainAssignment = {
    team_id:    string
    captain_id: string
}

type Draft  = {
    id?: string;
    name: string;
    owner?: string;
    captains: TeamCaptainAssignment[];
    available: string[];
    selections?: string[];
    format: string;
    rating_cutoffs: Record<string, number>;
    completedAt?: Date;
    draft_order_pattern: string;
}

type DraftOrderPattern = {
    name: string;
    description: string;
    example: number[][];
}
