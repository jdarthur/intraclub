import {combineReducers, configureStore} from '@reduxjs/toolkit';
import {mainApi} from "./api.js";
import authReducer from './auth.js';


const reducers = combineReducers({
    [mainApi.reducerPath]: mainApi.reducer,
    auth: authReducer,
})

export const store = configureStore({
    reducer: reducers,
    middleware: (getDefaultMiddleware) =>
        getDefaultMiddleware({}).concat(mainApi.middleware)
})

