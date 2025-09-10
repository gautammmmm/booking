import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import {
  Container,
  Paper,
  Typography,
  Box,
  CircularProgress,
  Alert,
  Grid,
  Card,
  CardContent,
  Button,
} from '@mui/material';
import { slotsAPI, servicesAPI } from '../services/api'; // Import from api.js

const PublicBooking = () => {
  const { businessId } = useParams();
  const [services, setServices] = useState([]);
  const [slots, setSlots] = useState([]);
  const [selectedService, setSelectedService] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadServices();
  }, [businessId]);

  const loadServices = async () => {
    try {
      setLoading(true);
      const response = await servicesAPI.getPublic(businessId); // Use the imported function
      setServices(response.data);
    } catch (error) {
      setError('Failed to load services');
    } finally {
      setLoading(false);
    }
  };

  const loadSlots = async (serviceId) => {
    try {
      setLoading(true);
      const response = await slotsAPI.getPublic(businessId, serviceId);
      setSlots(response.data);
      setSelectedService(serviceId);
    } catch (error) {
      setError('Failed to load available slots');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Paper elevation={3} sx={{ p: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom align="center">
          Book an Appointment
        </Typography>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <Typography variant="h6" gutterBottom>
          Select a Service:
        </Typography>

        <Grid container spacing={2} sx={{ mb: 4 }}>
          {services.map((service) => (
            <Grid item xs={12} sm={6} md={4} key={service.id}>
              <Card 
                sx={{ 
                  cursor: 'pointer',
                  bgcolor: selectedService === service.id ? 'primary.light' : 'background.paper'
                }}
                onClick={() => loadSlots(service.id)}
              >
                <CardContent>
                  <Typography variant="h6">{service.name}</Typography>
                  <Typography color="text.secondary">
                    {service.duration} minutes
                  </Typography>
                  {service.description && (
                    <Typography variant="body2" sx={{ mt: 1 }}>
                      {service.description}
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>

        {selectedService && (
          <>
            <Typography variant="h6" gutterBottom>
              Available Time Slots:
            </Typography>
            
            {loading ? (
              <Box display="flex" justifyContent="center" sx={{ py: 4 }}>
                <CircularProgress />
              </Box>
            ) : (
              <Grid container spacing={2}>
                {slots.map((slot) => (
                  <Grid item xs={12} sm={6} md={4} key={slot.id}>
                    <Card>
                      <CardContent>
                        <Typography variant="h6">
                          {new Date(slot.start_time).toLocaleDateString()}
                        </Typography>
                        <Typography>
                          {new Date(slot.start_time).toLocaleTimeString()} - 
                          {new Date(slot.end_time).toLocaleTimeString()}
                        </Typography>
                        <Button 
                          variant="contained" 
                          sx={{ mt: 2 }}
                          fullWidth
                        >
                          Book This Slot
                        </Button>
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            )}
          </>
        )}
      </Paper>
    </Container>
  );
};

export default PublicBooking;