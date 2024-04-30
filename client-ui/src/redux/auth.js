import {createSlice} from '@reduxjs/toolkit';
import {useMemo} from "react";
import {useSelector} from "react-redux";

export const IntraclubTokenKey = "intraclub-token"

const slice = createSlice({
    name: 'auth',
    initialState: {token: ""},
    reducers: {
        setCredentials: (state, {payload: {token}}) => {
            state.token = token;
            sessionStorage.setItem(IntraclubTokenKey, token);
        },
        logoutUser: (state) => {
            state.token = "";
            sessionStorage.removeItem(IntraclubTokenKey);
        }
    }
});

export const {setCredentials, logoutUser} = slice.actions;

export default slice.reducer;

export const useToken = () => {
    const auth = useSelector(selectCurrentAuth);

    return useMemo(() => {
        return auth.token;
    }, [auth]);
};

export const selectCurrentAuth = (state) => state.auth;
