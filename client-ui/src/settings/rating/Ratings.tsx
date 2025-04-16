import * as React from 'react';
import {useDeleteRatingMutation, useGetRatingsQuery} from "../../redux/api.js";
import {Card, Empty, Space,} from "antd";
import {RatingForm} from "./RatingForm";
import {LabeledValue} from "../../common/LabeledValue";
import {DeleteConfirm} from "../../common/DeleteConfirm";
import {Ellipsis} from "../../common/Ellipsis";

export function Ratings() {
    const {data} = useGetRatingsQuery()

    const ratings = data?.resource?.map((f: Rating) => (
        <OneRating key={f.id} id={f.id} user_id={f.user_id} name={f.name} description={f.description}/>
    ))

    return <div style={{display: "flex", flexWrap: "wrap", gap: "1em"}}>
        {ratings?.length ? ratings : <Empty/>}
        <div style={{height: "1em"}}/>
        <RatingForm/>
    </div>
}

export function OneRating({id, user_id, name, description}: Rating) {
    const [deleteRating] = useDeleteRatingMutation()
    const deleteSelf = (): string => {
        return deleteRating(id).then((res: { error: any; data: any; }) => (res.error ? res.error : ""));
    }

    const initialState: Rating = {
        id, user_id, name, description,
    }

    const title = <Ellipsis fullValue={name} maxLength={20}/>
    const editForm = <RatingForm RatingId={id} InitialState={initialState} Update/>
    const extra = <Space>
        {editForm}
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"rating"}/>
    </Space>

    return <Card title={title} style={{width: 250}} size={"small"} extra={extra}>
        <LabeledValue label={"Description"} value={<i>{description}</i>} vertical/>
    </Card>
}