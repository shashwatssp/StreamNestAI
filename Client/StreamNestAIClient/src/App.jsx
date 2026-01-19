import { useState, useEffect } from 'react'
import './App.css'
import Home from './components/home/Home';
import Recommended from './components/recommended/Recommended';
import Review from './components/review/Review';
import Header from './components/header/Header';
import Register from './components/register/Register';
import Login from './components/login/Login';
import Layout from './components/Layout';
import RequiredAuth from './components/RequiredAuth';
import axiosClient from './api/axiosConfig';
import useAuth from './hooks/useAuth';
import StreamMovie from './components/stream/StreamMovie';

import {Route, Routes, useNavigate} from 'react-router-dom'
import Explore from './components/explore/Explore';

// New feature components
import SocialFeatures from './components/social/SocialFeatures';
import AdminDashboard from './components/admin/AdminDashboard';
import AnalyticsDashboard from './components/analytics/AnalyticsDashboard';
import Recommendations from './components/recommendations/Recommendations';
import WatchlistManager from './components/watchlist/WatchlistManager';
import EnhancedProfile from './components/profile/EnhancedProfile';
import SystemMonitoring from './components/monitoring/SystemMonitoring';
import NaturalLanguageSearch from './components/search/NaturalLanguageSearch';
import ErrorBoundary from './components/ErrorBoundary';

function App() {

  const navigate = useNavigate();
  const { auth, setAuth } = useAuth();

  
  const updateMovieReview = (imdb_id) => {
      navigate(`/review/${imdb_id}`);
  };
   
  const handleLogout = async () => {

        try {
            const response = await axiosClient.post("/logout",{user_id: auth.user_id});
            console.log(response.data);
            setAuth(null);
           // localStorage.removeItem('user');
            console.log('User logged out');

        } catch (error) {
            console.error('Error logging out:', error);
        } 

    };

  return (
    <>
      <Header handleLogout = {handleLogout}/>
      <Routes path="/" element = {<Layout/>}>
        <Route path="/" element={<Home updateMovieReview={updateMovieReview}/>}></Route>
        <Route path="/register" element={<Register/>}></Route>
        <Route path="/login" element={<Login/>}></Route>
        <Route element = {<RequiredAuth/>}>
            <Route path="/recommended" element={<Recommended/>}></Route>
            <Route path="/explore" element={<Explore />} />
            <Route path="/review/:imdb_id" element={<Review/>}></Route>
            <Route path="/stream/:yt_id" element={<StreamMovie/>}></Route>
            
            {/* New feature routes with error boundaries */}
            <Route path="/social" element={
              <ErrorBoundary>
                <SocialFeatures />
              </ErrorBoundary>
            } />
            <Route path="/admin" element={
              <ErrorBoundary>
                <AdminDashboard />
              </ErrorBoundary>
            } />
            <Route path="/analytics" element={
              <ErrorBoundary>
                <AnalyticsDashboard />
              </ErrorBoundary>
            } />
            <Route path="/recommendations" element={
              <ErrorBoundary>
                <Recommendations />
              </ErrorBoundary>
            } />
            <Route path="/watchlist" element={
              <ErrorBoundary>
                <WatchlistManager />
              </ErrorBoundary>
            } />
            <Route path="/profile-enhanced" element={
              <ErrorBoundary>
                <EnhancedProfile />
              </ErrorBoundary>
            } />
            <Route path="/monitoring" element={
              <ErrorBoundary>
                <SystemMonitoring />
              </ErrorBoundary>
            } />
            <Route path="/search" element={
              <ErrorBoundary>
                <NaturalLanguageSearch />
              </ErrorBoundary>
            } />
        </Route>
      </Routes>

    </>
  )
}

export default App