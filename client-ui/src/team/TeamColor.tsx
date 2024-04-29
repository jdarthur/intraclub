import * as React from 'react';

export type TeamColorProps = {
    name: string,
    hex: string,
}

export function TeamColor({hex}: TeamColorProps) {
    return <div style={{backgroundColor: hex, width: 50, height: 50}}/>
}

type TeamNameAndColorProps = {
    name: string,
    color: TeamColorProps
}

export function TeamNameAndColor({name, color}: TeamNameAndColorProps) {
    return <span>
        <TeamColor name={name} hex={color.hex}/>
        {name}
    </span>
}