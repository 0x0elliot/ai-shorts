"use client"

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { siteConfig } from '@/app/siteConfig';

const CampaignHistory = ({ isOpen, onClose}) => {
  const [campaigns, setCampaigns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const accessToken = localStorage.getItem('accessToken');


  useEffect(() => {
    if (isOpen && accessToken) {
      fetchCampaigns();
    }
  }, [isOpen, accessToken]);

  const fetchCampaigns = async () => {
    try {
      const response = await fetch(`${siteConfig.baseApiUrl}/api/notification/private/notification-campaigns`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${accessToken}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch campaigns');
      }

      const data = await response.json();
      setCampaigns(data);
      setLoading(false);
    } catch (error) {
      console.error("Error fetching campaigns:", error);
      setError(error.message);
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center">
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg max-w-md w-full">
        <h2 className="text-2xl font-semibold text-gray-900 dark:text-gray-50 mb-4">Notification Campaigns</h2>
        
        {loading ? (
          <p className="text-gray-600 dark:text-gray-400">Loading campaigns...</p>
        ) : error ? (
          <p className="text-red-500">{error}</p>
        ) : (
          <div className="max-h-80 overflow-y-auto mb-4">
            {campaigns.length === 0 ? (
              <p className="text-gray-600 dark:text-gray-400">No campaigns found.</p>
            ) : (
              campaigns.map((campaign) => (
                <div key={campaign.id} className="bg-gray-100 dark:bg-gray-700 p-3 rounded-md mb-2">
                  <h3 className="font-medium text-gray-800 dark:text-gray-200">{campaign.name}</h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">{campaign.description}</p>
                </div>
              ))
            )}
          </div>
        )}
        
        <div className="flex justify-end">
          <Button
            className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400"
            onClick={onClose}
          >
            Close
          </Button>
        </div>
      </div>
    </div>
  );
};

export default CampaignHistory; 