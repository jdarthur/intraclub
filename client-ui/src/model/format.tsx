
type FormatLine = {
    player_1_rating: string
    player_2_rating: string
}

type Format = {
    id?: string
    user_id?: string
    name: string
    possible_ratings: string[]
    lines: FormatLine[]
}