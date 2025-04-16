import * as React from 'react';
import {useDeleteRatingMutation, useGetFormatsQuery, useGetRatingsQuery} from "../../redux/api.js";
import {Card, Empty, Space, Spin, Tag, Tooltip,} from "antd";
import {FormatForm} from "./FormatForm";
import {LabeledValue} from "../../common/LabeledValue";
import {DeleteConfirm} from "../../common/DeleteConfirm";
import {Ellipsis} from "../../common/Ellipsis";
import {InfoCircleOutlined} from "@ant-design/icons";

export function Formats() {
    const {data} = useGetFormatsQuery();

    const formats = data?.resource?.map((f: Format) => (
        <OneFormat key={f.id} id={f.id} user_id={f.user_id} name={f.name}
                   possible_ratings={f.possible_ratings} lines={f.lines}/>
    ))

    return <div style={{display: "flex", flexWrap: "wrap", gap: "1em"}}>
        {formats?.length ? formats : <Empty/>}
        <div style={{height: "1em"}}/>
        <FormatForm/>
    </div>
}

export function OneFormat({id, user_id, name, possible_ratings, lines}: Format) {
    const [deleteRating] = useDeleteRatingMutation()
    const deleteSelf = (): string => {
        return deleteRating(id).then((res: { error: any; data: any; }) => (res.error ? res.error : ""));
    }

    const initialState: Format = {
        id, user_id, name, possible_ratings, lines
    }

    const title = <Ellipsis fullValue={name} maxLength={20}/>
    const editForm = <FormatForm FormatId={id} InitialState={initialState} Update/>
    const extra = <Space>
        {editForm}
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"rating"}/>
    </Space>

    return <Card title={title} style={{width: 300}} size={"small"} extra={extra}>
        <LabeledValue label={"Ratings"} value={<RatingsTagDisplay IdList={possible_ratings}/>} vertical/>
        <LabeledValue label={"Lines"} value={<FormatLinesTagDisplay FormatLines={lines}/>} vertical/>
    </Card>
}


type StringIdsDisplayProps = {
    IdList: string[]
}

function RatingsTagDisplay({IdList}: StringIdsDisplayProps) {
    const tags = IdList?.map((ratingId: string) => {
        return <Tag>
            <RatingTooltip RatingId={ratingId}/>
        </Tag>
    })
    return <span style={{display: "flex", flexWrap: "wrap", gap: "0.5em 0em"}}>{tags}</span>
}

type RatingTooltipProps = {
    RatingId: string
}

function RatingTooltip({RatingId}: RatingTooltipProps) {
    const {data, isFetching} = useGetRatingsQuery()
    if (isFetching) {
        return <Spin/>
    }

    const f = getRatingFromList(RatingId, data?.resource)
    return <Space size={"small"}>
        {f?.name}
        <Tooltip title={f?.description}>
            <InfoCircleOutlined style={{cursor: "pointer"}}/>
        </Tooltip>
    </Space>
}

type FormatLinesTagDisplayProps = {
    FormatLines: FormatLine[]
}

function FormatLinesTagDisplay({FormatLines}: FormatLinesTagDisplayProps) {
    const tags = FormatLines?.map((line: FormatLine) => (<FormatLineDisplay FormatLine={line}/>))
    return <span style={{display: "flex", flexWrap: "wrap"}}>{tags}</span>
}


type FormatLineDisplayProps = {
    FormatLine: FormatLine
}

function FormatLineDisplay({FormatLine}: FormatLineDisplayProps) {
    return <span>
        <Tag> <RatingTooltip RatingId={FormatLine.player_1_rating}/></Tag>
         <span style={{marginRight: "0.5em"}}>&</span>
        <Tag><RatingTooltip RatingId={FormatLine.player_2_rating}/></Tag>
    </span>
}

function getRatingFromList(ratingId: string, all: Rating[]): Rating {
    return all.find(f => f.id === ratingId)
}