import {Card, Empty} from "antd";
import * as React from "react";
import {SkillInfo, SkillInfoRecord} from "./SkillInfo";
import {ExperienceForm} from "./ExperienceForm";
import {useGetSkillInfoQuery} from "../redux/api.js";


type ExperienceForUserProps = {
    UserId: string
}

export function ExperienceForUser({UserId}: ExperienceForUserProps) {

    const {data} = useGetSkillInfoQuery(UserId)
    const info: SkillInfoRecord[] = data?.resource

    let content: React.JSX.Element = <Empty/>
    if (info?.length > 0) {
        content = <div>
            {info.map((r: SkillInfoRecord) => (<SkillInfo line={r.line}
                                                          level={r.level}
                                                          captain={r.captain}
                                                          most_recent_year={r.most_recent_year}
                                                          league_type={r.league_type}
                                                          id={r.id}
                                                          key={r.id}/>))
            }
        </div>
    }

    return <Card title={"Experience"} style={{width: 500}}>
        {content}
        <ExperienceForm UserId={UserId}/>
    </Card>
}