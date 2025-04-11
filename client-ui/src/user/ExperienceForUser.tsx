import {Card, Empty} from "antd";
import * as React from "react";
import {SkillInfo, SkillInfoRecord} from "./SkillInfo";
import {ExperienceForm} from "./ExperienceForm";
import {useGetSkillInfoQuery, useWhoAmIQuery} from "../redux/api.js";
import {USER_PAGE_WIDTH} from "./UserPage";
import {useToken} from "../redux/auth";


type ExperienceForUserProps = {
    UserId: string
}

export function ExperienceForUser({UserId}: ExperienceForUserProps) {

    const token = useToken()
    const {data: whoami} = useWhoAmIQuery({}, {skip: !token})

    const {data} = useGetSkillInfoQuery(UserId)
    const info: SkillInfoRecord[] = data?.resource

    let content: React.JSX.Element = <Empty/>
    if (info?.length > 0) {
        content =
            <div style={{display: "flex", flexWrap: "wrap", flexDirection: "row", justifyContent: "space-evenly"}}>
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

    let extra = null
    if (whoami?.id == UserId) {
        extra = <ExperienceForm UserId={UserId}/>
    }

    return <Card title={"Experience"}
                 style={{width: USER_PAGE_WIDTH}}
                 styles={{body: {padding: "0.5em"}}}
                 extra={extra}>
        <div>
            {content}
        </div>
    </Card>
}