// storageService.ts
import { LocalStorage } from '@/types/storage';
import { currentUser } from '@clerk/nextjs';

type ServerResponse = {
  success: boolean;
  data?: any;  // You can be more specific with this type if you know the expected structure
  message?: string;
};


export const storageService = {
  async getItemFromServer(key: keyof LocalStorage): Promise<ServerResponse | null> {
    try {
      // const url = `${process.env.PUBLIC_BACKEND_API_URL}/api/sync/localstorage/?key=${encodeURIComponent(key as string)}`;
      const url = `http://localhost:8080/api/sync/localstorage/?key=${encodeURIComponent(key as string)}`;

      const sessionToken = localStorage.getItem('sessionToken'); // Adjust this line to wherever your session token is stored
      console.log('sessionToken:', sessionToken)

      const headers: HeadersInit = {
        'Content-Type': 'application/json'
      };

      if (sessionToken) {
        headers['Authorization'] = `Bearer ${sessionToken}`;
      }


      const options: RequestInit = {
        method: 'GET',
        headers: headers,
        mode: 'cors',
        credentials: 'include',
      };

      const response = await fetch(url, options);

      if (!response.ok) {
        // You can first attempt to decode the response as JSON, and then fall back to text if it fails.
        let errorMessage = 'Server responded with an error';
        try {
          const errorBody = await response.json();
          errorMessage = errorBody.message || `Server responded with ${response.status}`;
        } catch (jsonError) {
          errorMessage = await response.text(); // If response is not in JSON format
        }

        throw new Error(errorMessage);
      }

      // If the response is OK, we decode it from JSON
      const data: ServerResponse = await response.json();
      return data;

    } catch (error) {
      console.error(`Request failed: ${error}`);
      return null
    }
  },

  async setItemOnServer(key: keyof LocalStorage, value: LocalStorage[keyof LocalStorage]): Promise<string | null> {
    try {
      // const url = `${process.env.PUBLIC_BACKEND_API_URL}/api/sync/localstorage/?key=${encodeURIComponent(key as string)}`;
      const url = `http://localhost:8080/api/sync/localstorage/?key=${encodeURIComponent(key as string)}`;

      const headers: HeadersInit = {
        'Content-Type': 'application/json'
      };


      const options: RequestInit = {
        method: 'POST',
        headers: headers,
        mode: 'cors',
        body: value as string,
        credentials: 'include',
      };

      const response = await fetch(url, options);
      const responseClone = response.clone(); // Clone the response to read it multiple times

      if (!response.ok) {
        let errorMessage = 'Server responded with an error';
        try {
          const errorBody = await responseClone.json(); // Try to parse as JSON first
          errorMessage = errorBody.message || `Server responded with status code ${response.status}`;
        } catch (jsonError) {
          errorMessage = await responseClone.text(); // If response is not JSON, read as text
        }

        throw new Error(errorMessage);
      }

      // handle response logic, if needed
      const data = await response.json();
      // console.log(data);
      return JSON.stringify(data);  // or just `return data;` if you don't need to stringify the response

    } catch (error) {
      console.error('Failed to sync with server:', error);
      return null;
    }
  },


  async ClearConversationsFromServer(): Promise<string | null> {
    try {
      // const url = `${process.env.PUBLIC_BACKEND_API_URL}/api/sync/localstorage/?key=${encodeURIComponent(key as string)}`;
      const key = `clearConversations`
      const url = `http://localhost:8080/api/sync/localstorage/?key=${encodeURIComponent(key as string)}`;

      const headers: HeadersInit = {
        'Content-Type': 'application/json'
      };


      const options: RequestInit = {
        method: 'DELETE',
        headers: headers,
        mode: 'cors',
        credentials: 'include',
      };

      const response = await fetch(url, options);

      if (!response.ok) {
        // You can first attempt to decode the response as JSON, and then fall back to text if it fails.
        let errorMessage = 'Server responded with an error';
        try {
          const errorBody = await response.json();
          errorMessage = errorBody.message || `Server responded with ${response.status}`;
        } catch (jsonError) {
          errorMessage = await response.text(); // If response is not in JSON format
        }

        throw new Error(errorMessage);
      }

      // If the response is OK, we decode it from JSON
      const data = await response.json();
      return JSON.stringify(data);  // or just `return data;` if you don't need to stringify the response
    } catch (error) {
      console.error(`Request failed: ${error}`);
      return null;
    }
  },
};

