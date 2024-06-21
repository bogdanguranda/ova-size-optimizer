#!/bin/bash

ARCHIVES=("archives/infra-promstack-images-24.4.0-446.tar.gz" "archives/infra-images-24.4.0-446.tar.gz" "archives/aiops-images-24.4.0-69.tar.gz" "archives/secapp-images-24.4.0-247.tar.gz" "archives/appd-images-24.4.0-446.tar.gz")
LOG_FILE="performance-metrics.log"
TOTAL_START=$(date +%s)

echo "Starting performance measurements for multiple archives..." > $LOG_FILE

for ARCHIVE in "${ARCHIVES[@]}"; do
    echo "Processing $ARCHIVE" >> $LOG_FILE
    START_TIME=$(date +%s)
    
    /usr/bin/time -p make COMPRESSED_MULTI_ARCHIVE=$ARCHIVE 2>> $LOG_FILE &
    MAKE_PID=$!
    
    # Monitor the process with top
    echo "Monitoring performance for $ARCHIVE (PID: $MAKE_PID)" >> $LOG_FILE
    top -pid $MAKE_PID -stats pid,command,cpu,mem -l 1 >> $LOG_FILE
    
    wait $MAKE_PID
    END_TIME=$(date +%s)
    
    ELAPSED_TIME=$((END_TIME - START_TIME))
    echo "Time taken for $ARCHIVE: $ELAPSED_TIME seconds" >> $LOG_FILE
    echo "==============================================" >> $LOG_FILE
done

TOTAL_END=$(date +%s)
TOTAL_ELAPSED=$((TOTAL_END - TOTAL_START))
echo "Total time taken for all archives: $TOTAL_ELAPSED seconds" >> $LOG_FILE

echo "Performance measurements completed. Results saved to $LOG_FILE."
