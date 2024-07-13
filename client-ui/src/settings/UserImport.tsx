import {Button, Modal, Space, Tag, Tooltip, Upload, UploadFile, UploadProps} from "antd";
import * as React from "react";
import {User} from "./Users";
import {
    CheckCircleOutlined, CheckSquareOutlined,
    CloseSquareOutlined,
    ImportOutlined,
    InfoCircleOutlined,
    UploadOutlined,
    XOutlined
} from "@ant-design/icons";
import {useImportUsersMutation} from "../redux/api.js";
import {PreformattedText} from "../common/PreformattedText";
import {StepForm, StepFormStep} from "../common/StepForm";
import {SubmitResult} from "../common/CommonFormModal";

export function UserImport() {

    const [open, setOpen] = React.useState<boolean>(false);
    const [fileList, setFileList] = React.useState<UploadFile[]>([]);
    const [fileData, setFileData] = React.useState<string>("");


    const [doImport, result] = useImportUsersMutation();

    const importUsers = async (): Promise<SubmitResult> => {
        console.log("Importing users from data:", fileData)
        return await doImport(fileData)
    }

    const onChange = async (file: any) => {
        if (file?.file?.text) {
            const text = await file?.file?.text()
            setFileData(text)
        } else {
            setFileData("")
        }
    }

    const results = result?.data?.resource?.map((r: any) => <UserImportResult success={r.success} error={r.error}
                                                                              user={r.user} key={r.user.email}/>)

    const preview = <div style={{marginTop: "1em"}}>
        <h3>Preview:</h3>
        <PreformattedText text={fileData} maxHeight={400} lineNumbers/>
    </div>

    const props: UploadProps = {
        onRemove: (file) => {
            const index = fileList.indexOf(file);
            const newFileList = fileList.slice();
            newFileList.splice(index, 1);
            setFileList(newFileList);
        },
        beforeUpload: (file) => {
            setFileList([...fileList, file]);
            return false;
        },
        fileList,
    };

    const tooltip = <span>
        Import multiple users from a CSV file. The required column
        headers in the top row of the CSV are
        <Tag style={{fontFamily: "monospace"}}>
          First Name, Last Name, Email
        </Tag>
    </span>

    const modalTitle = <Space>
        <span> Import users from CSV </span>
        <Tooltip title={tooltip}>
            <InfoCircleOutlined/>
        </Tooltip>
    </Space>

    const uploadStep = <div style={{}}>
        <div style={{display: "flex"}}>
            <Upload {...props} accept={"text/csv"} onChange={onChange}>
                <Button icon={<ImportOutlined/>}>Select File</Button>
            </Upload>
        </div>
        {fileData ? preview : null}
    </div>


    const steps: StepFormStep[] = [
        {
            title: "Select CSV file",
            content: uploadStep,
            onNext: importUsers
        },
        {
            title: "View results",
            content: results
        },
    ]

    const stepFormSubmit = async (): Promise<SubmitResult> => {
        setOpen(false);
        return {
            data: null,
            error: null
        }
    }

    return <div>
        <Modal open={open} title={modalTitle} onCancel={() => setOpen(false)} onOk={importUsers} footer={null}>
            <StepForm steps={steps} onStepFormFinish={stepFormSubmit}/>
        </Modal>
        <Button onClick={() => setOpen(true)} type={"primary"} style={{"marginBottom": "1em"}}>
            <Space>
                <ImportOutlined/>
                Import users
            </Space>
        </Button>
    </div>
}

type UserImportProps = {
    success: boolean
    error: string
    user: User
}

function UserImportResult({success, error, user}: UserImportProps) {

    const color = success ? "#f6ffed" : "#ffccc7"
    const icon = success ? <CheckSquareOutlined/> : <CloseSquareOutlined/>

    const message = error ? error : `User ${user.first_name} ${user.last_name} imported successfully`

    const tooltip = <Tooltip title={message}>
        {icon}
    </Tooltip>

    return <div style={{background: color, display: "flex", justifyContent: "space-between", padding: 5}}>
        <span style={{marginLeft: "0.5em"}}>
            {user.first_name} {user.last_name} ({user.email})
        </span>
        <span style={{marginRight: "0.5em", cursor: "pointer"}}>
            {tooltip}
        </span>
    </div>
}
