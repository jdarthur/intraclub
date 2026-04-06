import React, {useState} from 'react';
import {Button, Checkbox, Empty, Form as AntForm, Form, Input, Modal, Space} from 'antd';
import type {CheckboxChangeEvent} from 'antd/es/checkbox';
import {PlusOutlined} from '@ant-design/icons';
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {useGetDraftOrderPatternsQuery, useGetSeasonsCompositeQuery} from "../redux/api.js";
import {DraftForm} from "./NewDraft";

type SeasonFormValues = {
    seasonName: string;
    description: string;
};

type SeasonsCompositeQueryParams = {
    as_player: boolean;
    as_commissioner: boolean;
}

export const Seasons: React.FC = () => {
    const [queryParams, setQueryParams] = useState<SeasonsCompositeQueryParams>({
        as_player: true,
        as_commissioner: true
    });

    const {data: seasons} = useGetSeasonsCompositeQuery(queryParams);

    const handleNewSeasonClick = () => {
        setIsModalVisible(true); // Open the modal when the button is clicked
    };

    return (
        <div>
            <NavigationBreadcrumb items={["My Seasons"]}/>

            <SeasonsFilters handleNewSeasonClick={handleNewSeasonClick} queryParams={queryParams}
                            setQueryParams={setQueryParams}/>

            <div style={{display: "flex", justifyContent: "flex-start"}}>
                <Empty description={"No seasons"}/>
            </div>

        </div>
    );
};

type SeasonsFilterProps = {
    queryParams: SeasonsCompositeQueryParams;
    setQueryParams: (queryParams: SeasonsCompositeQueryParams) => void;
    handleNewSeasonClick: () => void;
}

function SeasonsFilters({queryParams, setQueryParams, handleNewSeasonClick}: SeasonsFilterProps) {

    const handleMemberSeasonsChange = (e: CheckboxChangeEvent) => {
        const newValues = {...queryParams, as_player: e.target.checked};
        setQueryParams(newValues);
    };

    const handleCommissionerSeasonsChange = (e: CheckboxChangeEvent) => {
        const newValues = {...queryParams, as_commissioner: e.target.checked};
        setQueryParams(newValues);
    };

    return <Space
        style={{
            marginBottom: 16,
            borderRadius: 8,
            backgroundColor: '#ffffff',
            width: 'fit-content',
            padding: '8px 16px',
            display: 'flex',
            alignItems: 'center',
            gap: '16px'
        }}
    >
        <Space>
            <Checkbox
                checked={queryParams.as_player}
                onChange={handleMemberSeasonsChange}
            >
                Player
            </Checkbox>
            <Checkbox
                checked={queryParams.as_commissioner}
                onChange={handleCommissionerSeasonsChange}
            >
                Commissioner
            </Checkbox>
        </Space>
        <DraftForm />
    </Space>
}

